package randomtransaction

import (
	"math/rand"
	"time"

	"github.com/tendermint/tendermint/txgpartition"
)

var (
	MAXACCOUNTNUM   = 200000
	ACCOUNTSPLIT    = 4
	READNUM         = ACCOUNTSPLIT / 2
	ACCOUNTDURATION = MAXACCOUNTNUM / ACCOUNTSPLIT
)

type IDNode struct {
	RIDs                 []int
	WIDs                 []int
	id                   int64
	Father, Child        []txgpartition.TxNode
	Indegreee, OutDegree int
}

var _ txgpartition.TxNode = (*IDNode)(nil)

func NewIDNode(reads []int, writes []int, ID int64) *IDNode {
	return &IDNode{
		RIDs:   reads,
		WIDs:   writes,
		Father: make([]txgpartition.TxNode, 0),
		Child:  make([]txgpartition.TxNode, 0),
		id:     ID,
	}
}

func NewRandomIDNode(ID int64) *IDNode {
	rand.Seed(time.Now().UnixNano())
	accounts := make([]int, ACCOUNTSPLIT)
	mod := rand.Intn(5)
	for i := 0; i < ACCOUNTSPLIT; i++ {
		accounts[i] = ACCOUNTDURATION*i + rand.Intn(ACCOUNTDURATION/5)*mod
	}

	rand.Shuffle(ACCOUNTSPLIT, func(i, j int) {
		accounts[i], accounts[j] = accounts[j], accounts[i]
	})

	return NewIDNode(accounts[:READNUM], accounts[READNUM:], ID)
}

func (n *IDNode) ID() int64 {
	return n.id
}
func (n *IDNode) Less(n2 txgpartition.TxNode) bool {
	if e, ok := n2.(*IDNode); ok {
		return n.id < e.id
	} else {
		return false
	}
}
func (n *IDNode) Equal(n2 txgpartition.TxNode) bool {
	if e, ok := n2.(*IDNode); ok {
		return n.id == e.id
	} else {
		return false
	}
}
