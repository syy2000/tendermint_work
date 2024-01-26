package txgpartition

/*
	事务图划分所需要的所有图操作
	一个简单的例子参见randomSquare/square.go
*/

type TxNode interface {
	// 确定性地对交易进行排序
	// 可以直接对事务的ID进行排序
	// 1. 每个事务节点都需要存储自己的ID，ID可以是任意唯一的int值
	Less(TxNode) bool
	Equal(TxNode) bool
	ID() int64
}

type TxGraph interface {
	// node infomation and node action
	// 判断节点是事务节点还是区块节点
	// 2. 区块状态映射表中的区块也要被当作事务节点看待，有自己的ID，有自己的子节点（没有父节点）
	IsBlockNode(TxNode) bool
	// 节点的入度（区块节点没有入度，即入度为0）
	InDegree(TxNode) int
	// 节点的出度
	OutDegree(TxNode) int
	// 节点的入度减1
	// 3. 也就是说，事务节点还必须记录自己的入度和出度
	DecOutDegree(TxNode)
	// 下面两个接口是表明节点是否被访问过，是DFS需要的接口，暂时没有用，Visited可以默认输出false
	Visited(TxNode) bool
	Visit(TxNode)
	// 节点的ID
	// 可以直接返回TxNode.ID()
	NodeIndex(TxNode) int64

	// graph basic
	// 图中区块节点的数量
	BlockNodeNum() int
	// 图中事务节点的数量
	TxNodeNum() int

	// query for nodes
	// 遍历找到图中所有零入度节点
	FindZeroOutdegree() []TxNode
	// 节点的所有父节点；可以是乱序的
	// 4. 不可以遍历所有节点寻找父节点，可以将父节点的指针暂存在数据结构中
	QueryFather(TxNode) []TxNode
	// 节点的所有子节点；可以是乱序的
	QueryNodeChild(TxNode) []TxNode
}
