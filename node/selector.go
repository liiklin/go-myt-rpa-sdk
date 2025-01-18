package node

import (
	"errors"
	"fmt"
	"mytrpc/rpc"
	"time"
	"unsafe"
)

func NewSelector(client *rpc.Client) *Selector {
	proc, err := client.GetDLL().FindProc("newSelector")
	if err != nil {
		return nil
	}

	handle, _, _ := proc.Call(uintptr(client.GetHandle()))
	if handle == 0 {
		return nil
	}

	return &Selector{
		handle:    handle,
		rpcClient: client,
	}
}

func (s *Selector) FindOne(timeout time.Duration) (*Node, error) {
	if s.handle == 0 {
		return nil, errors.New("selector handle is invalid")
	}

	// 使用 findNodes 函数查找节点
	proc, err := s.rpcClient.GetDLL().FindProc("findNodes")
	if err != nil {
		return nil, fmt.Errorf("查找findNodes函数失败: %v", err)
	}

	nodesHandle, _, _ := proc.Call(
		s.handle,
		uintptr(1), // maxNode = 1
		uintptr(timeout.Milliseconds()),
	)

	if nodesHandle == 0 {
		return nil, nil
	}

	// 获取节点数量
	proc, err = s.rpcClient.GetDLL().FindProc("getNodesSize")
	if err != nil {
		return nil, fmt.Errorf("查找getNodesSize函数失败: %v", err)
	}

	size, _, _ := proc.Call(nodesHandle)
	if size == 0 {
		return nil, nil
	}

	// 获取第一个节点
	proc, err = s.rpcClient.GetDLL().FindProc("getNodeByIndex")
	if err != nil {
		return nil, fmt.Errorf("查找getNodeByIndex函数失败: %v", err)
	}

	nodeHandle, _, _ := proc.Call(nodesHandle, 0)
	if nodeHandle == 0 {
		return nil, nil
	}

	// 清除查询条件
	if proc, err := s.rpcClient.GetDLL().FindProc("clearSelector"); err == nil {
		proc.Call(s.handle)
	}

	return &Node{
		handle:    nodeHandle,
		rpcClient: s.rpcClient,
	}, nil
}

// AddTextQuery 添加文本查询条件
func (s *Selector) AddTextQuery(text string) {
	if s.handle == 0 {
		return
	}

	proc, err := s.rpcClient.GetDLL().FindProc("TextContainWith")
	if err != nil {
		return
	}

	textBytes := []byte(text + "\x00")
	proc.Call(
		s.handle,
		uintptr(unsafe.Pointer(&textBytes[0])),
	)
}

// 其他查询条件方法...
