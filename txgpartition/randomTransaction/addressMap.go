package randomtransaction

import (
	"math/rand"
	"time"

	"github.com/tendermint/tendermint/txgpartition"
)

type AddressMap struct {
	RL map[int][]*IDNode
	WL map[int][]*IDNode

	blockNodeNum, txNodeNum int
	nodes                   []*IDNode
	edgeNum                 int
	execTime                time.Duration
}

var _ txgpartition.TxGraph = (*AddressMap)(nil)

func NewAddressMap(blockNodeNum, txNodeNum int) *AddressMap {
	return &AddressMap{
		RL:           make(map[int][]*IDNode, MAXACCOUNTNUM),
		WL:           make(map[int][]*IDNode, MAXACCOUNTNUM),
		blockNodeNum: blockNodeNum,
		txNodeNum:    txNodeNum,
		nodes:        make([]*IDNode, blockNodeNum+txNodeNum),
	}
}

func (am *AddressMap) Init() {
	avaNum := MAXACCOUNTNUM / am.blockNodeNum
	NumberList := make([]int, MAXACCOUNTNUM)
	for i := 0; i < MAXACCOUNTNUM; i++ {
		NumberList[i] = i
	}
	rand.Shuffle(MAXACCOUNTNUM, func(i, j int) {
		NumberList[i], NumberList[j] = NumberList[j], NumberList[i]
	})
	start := time.Now()
	for i := 0; i < am.blockNodeNum; i++ {
		u := NewIDNode(nil, NumberList[i*avaNum:(i+1)*avaNum], int64(i))
		am.nodes[i] = u
		am.AddNode(u)
	}
	for i := int64(am.blockNodeNum); i < int64(am.blockNodeNum+am.txNodeNum); i++ {
		u := NewRandomIDNode(i)
		am.nodes[i] = u
		am.AddNode(u)
	}
	am.execTime = time.Since(start)
}

func (am *AddressMap) ConstructTime() time.Duration {
	return am.execTime
}

func (am *AddressMap) AddRead(n *IDNode, addr int) {
	if u, ok := am.WL[addr]; ok && len(u) > 0 {
		am.BuildRelation(n, u)
	}
	am.RL[addr] = append(am.RL[addr], n)
}
func (am *AddressMap) AddWrite(n *IDNode, addr int) {
	if u, ok := am.RL[addr]; ok && len(u) > 0 {
		am.BuildRelation(n, u)
	} else if u, ok := am.WL[addr]; ok && len(u) > 0 {
		am.BuildRelation(n, u)
	}
	am.RL[addr] = nil
	am.WL[addr] = []*IDNode{n}
}
func (am *AddressMap) AddNode(n *IDNode) {
	for _, u := range n.RIDs {
		am.AddRead(n, u)
	}
	for _, u := range n.WIDs {
		am.AddWrite(n, u)
	}
}
func (am *AddressMap) BuildRelation(n *IDNode, u []*IDNode) {
	for _, father := range u {
		if _, ok := n.Father[father.ID()]; !ok {
			n.OutDegree++
			father.Indegreee++
			n.Father[father.ID()] = father
			father.Child[n.ID()] = n
			am.edgeNum++
		}
	}
}

// ========================== TxGraph ===========================================
func (am *AddressMap) IsBlockNode(n txgpartition.TxNode) bool {
	return n.ID() < int64(am.blockNodeNum)
}
func (am *AddressMap) InDegree(n txgpartition.TxNode) int {
	return MustIDNode(n).Indegreee
}
func (am *AddressMap) OutDegree(n txgpartition.TxNode) int {
	return MustIDNode(n).OutDegree
}
func (am *AddressMap) DecOutDegree(n txgpartition.TxNode) {
	u := MustIDNode(n)
	u.OutDegree--
}
func (am *AddressMap) Visit(n txgpartition.TxNode) {
	// Nothing To Do
}
func (am *AddressMap) Visited(n txgpartition.TxNode) bool {
	return false
}
func (am *AddressMap) NodeIndex(n txgpartition.TxNode) int64 {
	return n.ID()
}
func (am *AddressMap) BlockNodeNum() int {
	return am.blockNodeNum
}
func (am *AddressMap) TxNodeNum() int {
	return am.txNodeNum
}
func (am *AddressMap) FindZeroOutdegree() (out []txgpartition.TxNode) {
	for _, n := range am.nodes {
		if n.OutDegree == 0 {
			out = append(out, n)
		}
	}
	return
}
func (am *AddressMap) QueryFather(n txgpartition.TxNode) map[int64]txgpartition.TxNode {
	u := MustIDNode(n)
	return u.Father
}
func (am *AddressMap) QueryNodeChild(n txgpartition.TxNode) map[int64]txgpartition.TxNode {
	u := MustIDNode(n)
	return u.Child
}

// ==============================================================================
func MustIDNode(n txgpartition.TxNode) *IDNode {
	if e, ok := n.(*IDNode); ok {
		return e
	} else {
		panic("this should not happen")
	}
}
