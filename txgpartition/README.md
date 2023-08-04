# 一、区块状态映射表部分
## 1. 概述
区块状态映射表(BlockStatusMappingTable)使用键值存储，存储了地址主键(string)对区块号(int64)的映射。实现了键值存储所必须的接口。使用以下代码可以导入区块状态映射表部分：
```
import (
"github.com/tendermint/tendermint/txpartition/statustable"
)
```
相应目录下，区块状态映射表的实现在文件BlockStatusMappingTable.go中。
## 2. 区块状态映射表提供的接口
区块状态映射表提供键值查询和存储所必须的接口，以及一个计算表哈希的接口。

|接口名称|  输入  |输出|功能描述|
|:-----:|:-----------:|:--:|:-----|
|Get|key string| value int64, ok bool|根据所给的string类型的key在区块状态映射表中查询最近一次对该地址进行修改的区块号(value)。如果不存在所查询的主键，则ok返回值为false。
|Set|key string, value int64|ok  bool|在区块状态映射表中记录key对value的映射；如果映射已存在，则覆盖原有记录。返回值记录修改是否成功。现阶段，如果返回值ok的值为false，可以认为是系统出现故障（内存溢出、栈溢出等），应立刻停止程序。
|Clear| 无|无|清空区块状态映射表
|Hash|无|[]byte|计算区块状态映射表的哈希值。该过程需要遍历一遍区块状态映射表中所有数据，因此耗时较长。

## 3. 区块装填映射表的创建
假设已经完成了1节中区块状态映射表的引入，使用以下函数可以创建一个指向区块状态映射表的指针：

`func NewBlockStatusMappingTable(tableType int8, options []func(Table)) *BlockStatusMappingTable`

### (1) 参数解读
#### ***tableType***：指定区块状态映射表的实现方式，有三个可选参数：

①  UseOrderedMap：

有序映射，分别由链表与map映射组成，通过链表记录区块状态到达的时间。查询直接通过map映射完成，插入需要额外在链表尾部添加一个记录，哈希的计算直接通过遍历链表实现。

***哈希值的随机性***：由于没有确定性排序，计算哈希的顺序等价于键到达的顺序，是否能够投入使用需要考虑共识的实现。

②  UseSimpleMap：

最简单的map映射。计算哈希时，需要首先将主键进行排序。

③  UseMPTree：

使用默克尔帕特里夏树实现的键值存储与哈希计算。数据较少时，查询与写入效率与使用map映射相当；万至百万数量级时，树的深度显著增加，查询速度降低。

***目前的测试中，在百万级别的数据集上，MPTree的哈希计算速度相对SimpleMap没有显著优势。***

④ UseSafeSimpleMap:

使用`sync.Map`实现的SimpleMap，可以进行并发安全的读写。注意，<font color="Red"> 这并不意味着hash()、clear()、set()三个接口之间可以并发</font>，因为后二者可能导致前者的哈希计算出现错误（脏读、不可重复度、幻读，因为hash并不是sync.Map的内置功能，而是通过使用get遍历表单实现的）。

#### ***options***：可选参数

目前仅提供了一个可选参数，即OnlyUseHashOptions。区块状态映射表使用MPTree实现时，启用该功能，可以对主键计算哈希值后进行存储，以确保数据分布的随机性，降低树的深度，该过程对用户透明。实验中，启用该功能，在万级别数据下，查询与写入速度提升了一倍以上，建议开启此功能。

### (2) 实例
创建一个OrderedMap实现的区块状态映射表：

```go
// 以下两条命令可以创建相同的区块状态映射表
t1 := statustable.NewBlockStatusMappingTable(0,nil)
t2 := statustable.NewBlockStatusMappingTable(statustable.UseOrderedMap,nil)
```
创建一个SimpleMap实现的区块状态映射表：
```go
// 以下两条命令可以创建相同的区块状态映射表
t1 := statustable.NewBlockStatusMappingTable(1,nil)
t2 := statustable.NewBlockStatusMappingTable(statustable.UseSimpleMap,nil)
```
创建一个MPTree实现的区块状态映射表：
```go
// 以下两条命令可以创建相同的区块状态映射表
t1 := statustable.NewBlockStatusMappingTable(2,nil)
t2 := statustable.NewBlockStatusMappingTable(statustable.UseMPTree,nil)
```
创建一个SimpleMap实现的**线程安全的**区块状态映射表：
```go
// 以下两条命令可以创建相同的区块状态映射表
t1 := statustable.NewBlockStatusMappingTable(3,nil)
t2 := statustable.NewBlockStatusMappingTable(statustable.UseSafeSimple,nil)
```
***<font color="Red">注意，即使是线程安全的区块状态映射表，clear()方法和hash()方法也是不可以同时调用的。</font>***
### (3) 测试

部分的测试代码可以在main_test.go中找到。

# 二、事务图划分部分

待完善

