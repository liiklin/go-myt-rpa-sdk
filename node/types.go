package node

import "mytrpc/rpc"

// Rect 表示节点的位置信息
type Rect struct {
	Left, Top, Right, Bottom int
}

// Node 表示一个UI节点
type Node struct {
	handle    uintptr
	rpcClient *rpc.Client
}

// Selector 用于查找节点
type Selector struct {
	handle    uintptr
	rpcClient *rpc.Client
}
