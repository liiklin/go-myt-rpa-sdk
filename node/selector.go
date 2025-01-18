package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"mytrpc/rpc"
	"time"
	"unsafe"
)

type Node struct {
	ID        string
	Text      string
	Class     string
	Enabled   bool
	Bounds    Rect
	Children  []*Node
	rpcClient *rpc.Client
}

type Rect struct {
	Left, Top, Right, Bottom int
}

type Selector struct {
	rpcClient *rpc.Client
	queries   []Query
}

type Query struct {
	Type  string
	Value interface{}
}

func NewSelector(client *rpc.Client) *Selector {
	return &Selector{
		rpcClient: client,
	}
}

func (s *Selector) AddTextQuery(text string) {
	s.queries = append(s.queries, Query{
		Type:  "text",
		Value: text,
	})
}

func (s *Selector) AddClassQuery(class string) {
	s.queries = append(s.queries, Query{
		Type:  "class",
		Value: class,
	})
}

func (s *Selector) AddEnabledQuery(enabled bool) {
	s.queries = append(s.queries, Query{
		Type:  "enabled",
		Value: enabled,
	})
}

func (s *Selector) FindOne(timeout time.Duration) (*Node, error) {
	// 创建选择器
	proc, err := s.rpcClient.GetDLL().FindProc("create_selector")
	if err != nil {
		return nil, fmt.Errorf("查找create_selector函数失败: %v", err)
	}

	ret, _, err := proc.Call(uintptr(s.rpcClient.GetHandle()))
	if ret == 0 {
		return nil, errors.New("创建选择器失败")
	}

	// 添加查询条件
	for _, query := range s.queries {
		var funcName string
		switch query.Type {
		case "text":
			funcName = "addQuery_TextContainWith"
		case "class":
			funcName = "addQuery_ClzEqual"
		case "enabled":
			funcName = "addQuery_Enable"
		}

		proc, err = s.rpcClient.GetDLL().FindProc(funcName)
		if err != nil {
			return nil, fmt.Errorf("查找%s函数失败: %v", funcName, err)
		}

		queryBytes := []byte(fmt.Sprintf("%v\x00", query.Value))
		ret, _, err = proc.Call(
			ret,
			uintptr(unsafe.Pointer(&queryBytes[0])),
		)
		if ret == 0 {
			return nil, fmt.Errorf("添加查询条件失败: %v", query)
		}
	}

	// 执行查询
	proc, err = s.rpcClient.GetDLL().FindProc("execQueryOne")
	if err != nil {
		return nil, fmt.Errorf("查找execQueryOne函数失败: %v", err)
	}

	ret, _, err = proc.Call(
		ret,
		uintptr(timeout.Milliseconds()),
	)

	if ret == 0 {
		return nil, nil
	}

	node := &Node{
		ID:        "",
		rpcClient: s.rpcClient,
	}

	return node, nil
}

func (n *Node) Click() error {
	proc, err := n.rpcClient.GetDLL().FindProc("clickNode")
	if err != nil {
		return fmt.Errorf("查找clickNode函数失败: %v", err)
	}

	ret, _, err := proc.Call(uintptr(unsafe.Pointer(&[]byte(n.ID)[0])))
	if ret == 0 {
		return errors.New("点击节点失败")
	}

	return nil
}

func (n *Node) GetJSON() (string, error) {
	data, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (n *Node) GetText() string {
	return n.Text
}

func (n *Node) GetBounds() Rect {
	return n.Bounds
}
