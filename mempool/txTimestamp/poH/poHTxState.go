package poH

import (
	"bytes"
	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/service"
	"github.com/tendermint/tendermint/libs/sync"
	"github.com/tendermint/tendermint/p2p"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types"
	"io"
	"sort"
)

const (
	MessageChanMax         = 1000
	TxWithTimestampChanMax = 100000
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
	PoHValidatorMap map[p2p.ID]*PoHValidator
	ValidatorMap    map[p2p.ID]*types.Validator
	cache           map[p2p.ID]*MessageCache
	mempool         *PoHMempool
	gen             *PoHGenerator
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
}

func NewPoHTxState(
	m *PoHMempool,
	gen *PoHGenerator,
	privKey crypto.PrivKey,
	pubKey crypto.PubKey,
	address crypto.Address,
) *PoHTxState {
	s := new(PoHTxState)
	s.mempool = m
	s.gen = gen
	s.PoHValidatorMap = make(map[p2p.ID]*PoHValidator)
	s.ValidatorMap = make(map[p2p.ID]*types.Validator)
	s.cache = make(map[p2p.ID]*MessageCache)
	s.MessageChan = make(chan *types.TxMessage, MessageChanMax)
	s.TxWithTimestampChan = make(chan types.TxWithTimestamp, TxWithTimestampChanMax)
	s.BaseService = *service.NewBaseService(nil, "timestampCenter", s)

	s.privKey = privKey
	s.pubKey = pubKey
	s.address = address

	s.OutPoHBlockPartSetChan = make(chan *types.PoHBlockPartSet, MessageChanMax)
	return s
}

func (s *PoHTxState) AddValidator(validator *types.Validator) bool {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	peerID := p2p.ID(validator.Address)
	_, ok := s.PoHValidatorMap[peerID]
	if ok {
		return true
	}
	v := NewPoHValidator(s.Logger)
	s.PoHValidatorMap[peerID] = v
	s.ValidatorMap[peerID] = validator
	s.cache[peerID] = NewMessageCache()
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
	for _, v := range s.PoHValidatorMap {
		v.SetSeed(seed)
	}
	s.mempool.SetSeed(seed)
	s.Height = seed.Height
	s.Round = seed.Round
	return s.gen.SetSeed(seed)
}

// TODO 错误处理
func (s *PoHTxState) OnStart() error {
	go func() error {
		for {
			select {
			case <-s.Quit():
				return nil
			case <-s.mempool.CreateBlockChan:
				s.handleCreateTxBlock()
			case mes := <-s.MessageChan:
				src := mes.Src
				switch d := mes.Data.(type) {
				case *types.PoHBlockPart:
					f, err := s.handleBlockPart(src, d)
					if err != nil {
						// TODO 需要细化
					}
					if !f {
						// TODO 需要细化
					}
				default:
					// TODO 出错
				}
			}
		}
	}()
	return nil
}

func (s *PoHTxState) handleCreateTxBlock() {
	txs := s.mempool.GetTimestamps()
	b := new(types.PoHBlock)
	b.Height = s.Height
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
	ps := types.NewPoHBlockPartSetFromData(bz, BlockPartSizeBytes, pb.Height)
	s.OutPoHBlockPartSetChan <- ps
	// ps.Total()
	// 这里是否需要转换为part输出？
}

func (s *PoHTxState) hasValidator(id p2p.ID) bool {
	_, ok := s.PoHValidatorMap[id]
	return ok
}

// TODO 定义错误类型
func (s *PoHTxState) handleBlockPart(src p2p.ID, p *types.PoHBlockPart) (bool, error) {
	if !s.hasValidator(src) {
		return false, &types.TimestampNormalError{}
	}
	height := p.Height
	if height < s.PoHValidatorMap[src].Height {
		return false, &types.TimestampNormalError{}
	}
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
			// r.Logger.Info("区块完整了", "height", p.Height, "total", p.Total, "index", p.Index)
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
		f, err := s.handleBlock(src, b)
		if !f || err != nil {
			return f, err
		}
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
	// 输出，需要改timestamp到tx，或者把Tx结构定下来也可以
	for _, tx := range b.PoHTimestamps {
		s.TxWithTimestampChan <- &types.Tx{
			OriginTx: tx.Message,
			TxTimehash: tx,
		}
	}
	return true, nil
}
