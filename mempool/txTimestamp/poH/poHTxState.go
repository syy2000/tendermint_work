package poH

import (
	"bytes"

	"io"
	"sort"

	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/service"
	"github.com/tendermint/tendermint/libs/sync"
	"github.com/tendermint/tendermint/p2p"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types"
)

const (
	MessageChanMax         = 1000000
	TxWithTimestampChanMax = 1000000
	BlockPartSizeBytes     = 65536
)

type MessageCache struct {
	PartSets    map[int64]*types.PoHBlockPartSet
	HeightSlice []int64
}

func NewMessageCache() *MessageCache {
	c := new(MessageCache)
	c.PartSets = make(map[int64]*types.PoHBlockPartSet)
	c.HeightSlice = make([]int64, 0)
	return c
}

type PoHTxState struct {
	PoHValidatorMap     map[p2p.ID]*PoHValidator
	ValidatorMap        map[p2p.ID]*types.Validator
	cache               map[p2p.ID]*MessageCache
	PoHValidatorTimeMap map[p2p.ID]int64

	lockMap map[p2p.ID]*sync.RWMutex

	blockChanMap map[p2p.ID]chan *types.PoHBlock

	mempool *PoHMempool
	gen     *PoHGenerator
	// 来自其他节点的消息输入
	MessageChan chan *types.TxMessage

	// 返回其他来自节点的tx
	TxWithTimestampChan chan types.TxWithTimestamp

	// 输出到reactor
	OutPoHBlockPartSetChan chan *types.PoHBlockPartSet

	seed   *types.Seed
	Height int64
	Round  int64

	privKey crypto.PrivKey
	pubKey  crypto.PubKey
	address crypto.Address

	mtx sync.Mutex
	service.BaseService

	txDone *TxDone

	Logger log.Logger
}

func NewPoHTxState(
	m *PoHMempool,
	gen *PoHGenerator,
	privKey crypto.PrivKey,
	pubKey crypto.PubKey,
	address crypto.Address,
	logger log.Logger,
) *PoHTxState {
	s := new(PoHTxState)
	s.mempool = m
	s.gen = gen
	s.PoHValidatorMap = make(map[p2p.ID]*PoHValidator)
	s.ValidatorMap = make(map[p2p.ID]*types.Validator)
	s.PoHValidatorTimeMap = make(map[p2p.ID]int64)
	s.cache = make(map[p2p.ID]*MessageCache)
	s.MessageChan = make(chan *types.TxMessage, MessageChanMax)
	s.TxWithTimestampChan = make(chan types.TxWithTimestamp, TxWithTimestampChanMax)
	s.BaseService = *service.NewBaseService(nil, "timestampCenter", s)

	s.privKey = privKey
	s.pubKey = pubKey
	s.address = address

	s.OutPoHBlockPartSetChan = make(chan *types.PoHBlockPartSet, MessageChanMax)

	s.txDone = gen.txDone

	s.Logger = logger

	s.lockMap = make(map[p2p.ID]*sync.RWMutex)
	s.blockChanMap = make(map[p2p.ID]chan *types.PoHBlock)
	return s
}

func (s *PoHTxState) AddValidator(validator *types.Validator) bool {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	peerID := p2p.ID(validator.Address)
	// peerID := p2p.ID(hex.EncodeToString(validator.Address))
	// s.Logger.Info("peerID", peerID, "Address", validator.Address)
	_, ok := s.PoHValidatorMap[peerID]
	if ok {
		return true
	}
	v := NewPoHValidator(s.Logger)
	s.PoHValidatorMap[peerID] = v
	s.ValidatorMap[peerID] = validator
	s.cache[peerID] = NewMessageCache()
	s.lockMap[peerID] = &sync.RWMutex{}
	s.blockChanMap[peerID] = make(chan *types.PoHBlock, 1000)
	go func(id p2p.ID) {
		ch := s.blockChanMap[id]
		for {
			select {
			case b := <-ch:
				s.handleBlock(p2p.ID(b.Address), b)
			}
		}
	}(peerID)
	// s.PoHValidatorTimeMap[peerID] = v.GetNowTimestamp().GetTimestamp()
	return true
}

func (s *PoHTxState) RemoveValidator(peerID p2p.ID) bool {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	_, ok := s.PoHValidatorMap[peerID]
	if !ok {
		return true
	}
	delete(s.PoHValidatorMap, peerID)
	delete(s.ValidatorMap, peerID)
	delete(s.cache, peerID)
	delete(s.PoHValidatorTimeMap, peerID)
	delete(s.blockChanMap, peerID)
	return true
}

func (s *PoHTxState) GetTxChan() chan types.TxWithTimestamp {
	return s.TxWithTimestampChan
}

func (s *PoHTxState) AddMessage(message *types.TxMessage) {
	s.MessageChan <- message
}

func (s *PoHTxState) SetSeed(seed *types.Seed) bool {
	s.seed = seed
	for id, v := range s.PoHValidatorMap {
		v.SetSeed(seed)
		s.PoHValidatorTimeMap[id] = seed.Round
	}
	s.mempool.SetSeed(seed)
	s.Height = seed.Height
	s.Round = seed.Round
	return s.gen.SetSeed(seed)
}

// TODO 错误处理
func (s *PoHTxState) OnStart() error {
	s.Logger.Info("txState 正在执行")
	go func() {
		for {
			select {
			case <-s.mempool.CreateBlockChan:
				s.handleCreateTxBlock()
			}
		}
	}()

	go func() error {
		for {
			select {
			case <-s.Quit():
				return nil
			case mes := <-s.MessageChan:
				src := mes.Src
				switch d := mes.Data.(type) {
				case *types.PoHBlockPart:
					// f, err := s.handleBlockPart(src, d)
					go s.getLock(src, d)
					// if err != nil {
					// 	// TODO 需要细化
					// }
					// if !f {
					// 	// TODO 需要细化
					// }
				default:
					// TODO 出错
				}
			}
		}
	}()
	return nil
}

func (s *PoHTxState) handleCreateTxBlock() {
	// s.Logger.Info("尝试发送poh区块", "height", s.Height)
	txs := s.mempool.GetTimestamps()
	flag := false
	for _, tx := range txs {
		if len(tx.Message) != 0 {
			flag = true
		}
	}
	if flag {
		// s.Logger.Info("创建的该区块不为空", "height", s.Height, "address", s.address)
	}
	b := new(types.PoHBlock)
	b.Height = s.Height
	s.Height++
	b.PoHTimestamps = txs
	// 签名
	b.Address = s.address
	err := b.Sign(s.privKey)
	if err != nil {
		// 出错
		return
	}
	// 拆分成part发送
	pb := b.ToProto()
	bz, err := proto.Marshal(pb)
	if err != nil {
		s.Logger.Error("区块拆分出错")
		return
	}
	ps := types.NewPoHBlockPartSetFromData(bz, BlockPartSizeBytes, pb.Height, b.Address)
	s.OutPoHBlockPartSetChan <- ps
	// s.Logger.Info("放入reactor")
	// ps.Total()
	// 这里是否需要转换为part输出？
}

func (s *PoHTxState) hasValidator(id p2p.ID) bool {
	_, ok := s.PoHValidatorMap[id]
	return ok
}

// TODO 定义错误类型
func (s *PoHTxState) handleBlockPart(src p2p.ID, p *types.PoHBlockPart) (bool, error) {
	// s.Logger.Info("我成功收到了消息", "address", p.Address, "height", p.Height)
	src = p2p.ID(p.Address)
	if !s.hasValidator(p2p.ID(p.Address)) {
		return false, &types.TimestampNormalError{}
	}
	// s.Logger.Info("有该验证者", "address", p.Address, "height", p.Height)
	height := p.Height
	if height < s.PoHValidatorMap[src].Height {
		return false, &types.TimestampNormalError{}
	}
	// s.Logger.Info("height正常", "address", p.Address, "height", p.Height)
	c := s.cache[src]
	ps, ok := c.PartSets[height]
	if !ok {
		ps = types.NewPoHBlockPartSetFromPart(p)
		c.PartSets[height] = ps
		c.HeightSlice = append(c.HeightSlice, height)
		sort.Slice(c.HeightSlice, func(i, j int) bool {
			return c.HeightSlice[i] < c.HeightSlice[j]
		})
	}
	ps.AddPart(p)
	if c.HeightSlice[0] != s.PoHValidatorMap[src].Height {
		return false, nil
	}

	res := make([]*types.PoHBlockPartSet, 0)
	index := 0
	for i, h := range c.HeightSlice {
		ps = c.PartSets[h]
		if ps.IsComplete() {
			// s.Logger.Info("区块完整了", "height", p.Height, "total", p.Total, "index", p.Index, "address", p.Address)
			res = append(res, ps)
			index = i + 1
			delete(c.PartSets, h)
		} else {
			break
		}
	}

	c.HeightSlice = c.HeightSlice[index:]
	for _, ps := range res {
		bz, err := io.ReadAll(ps.GetReader())
		if err != nil {
			return false, err
		}
		pbb := new(tmproto.PoHBlock)
		err = proto.Unmarshal(bz, pbb)
		if err != nil {
			return false, err
		}
		b := types.NewPoHBlockFromProto(pbb)
		// f, err := s.handleBlock(src, b)
		s.blockChanMap[src] <- b
		// if !f || err != nil {
		// 	return f, err
		// }
	}
	return true, nil
}

// TODO 验证签名、细化错误、输出
func (s *PoHTxState) handleBlock(src p2p.ID, b *types.PoHBlock) (bool, error) {
	if src != p2p.ID(b.Address) {
		return false, &types.ErrNotEnoughVotingPowerSigned{}
	}
	if !bytes.Equal(b.Address, s.ValidatorMap[src].Address) {
		return false, &types.ErrNotEnoughVotingPowerSigned{}
	}
	if !b.VerifySignature(s.ValidatorMap[src].PubKey) {
		return false, &types.ErrNotEnoughVotingPowerSigned{}
	}
	v := s.PoHValidatorMap[src]
	flag := v.Validate(b)
	if !flag {
		return false, &types.TimestampNormalError{}
	}
	// s.Logger.Info("我收到了区块 ", "address", b.Address, "height", b.Height, "time", v.lastTimestamp.GetTimestamp())
	// 输出，需要改timestamp到tx，或者把Tx结构定下来也可以
	for _, tx := range b.PoHTimestamps {
		memTx := &types.MemTx{
			OriginTx: types.Tx{
				OriginTx:   tx.Message,
				TxTimehash: tx,
			},
			TxTimehash: tx,
		}
		memTx.SetCallBack(func() {
			s.txDone.Done(memTx)
		})
		if len(tx.Message) != 0 {
			// s.Logger.Info("非空", "tx", tx)
			s.txDone.AddTxToTxDone(memTx)
			s.TxWithTimestampChan <- memTx
		}
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.PoHValidatorTimeMap[src] = v.lastTimestamp.GetTimestamp()
	return true, nil
}

func (s *PoHTxState) GetNowTimestamp2() int64 {
	s.Logger.Info("正在取出时间 state")
	s.mtx.Lock()
	defer s.mtx.Unlock()
	now := s.mempool.GetNowTimestamp()
	s.Logger.Info("正在取出时间 gen", "t", now)
	for _, v := range s.PoHValidatorTimeMap {
		t := v
		if now > t {
			now = t
		}
		s.Logger.Info("正在取出时间 v", "t", now)
	}
	s.Logger.Info("正在取出时间 after v", "t", now)
	t, ok := s.txDone.GetNowMin()
	if ok {
		if now > t {
			now = t
		}
		s.Logger.Info("正在取出时间 txDone", "t", now)
	}
	return now
}

func (s *PoHTxState) getLock(src p2p.ID, p *types.PoHBlockPart) (bool, error) {
	id := p2p.ID(p.Address)
	s.lockMap[id].Lock()
	defer s.lockMap[id].Unlock()
	return s.handleBlockPart(src, p)
}
