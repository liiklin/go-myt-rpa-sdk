package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"unsafe"
)

// Node 的方法实现
func (n *Node) Click() error {
	if n.handle == 0 {
		return errors.New("node handle is invalid")
	}

	proc, err := n.rpcClient.GetDLL().FindProc("clickNode")
	if err != nil {
		return fmt.Errorf("查找clickNode函数失败: %v", err)
	}

	ret, _, _ := proc.Call(n.handle)
	if ret == 0 {
		return errors.New("点击节点失败")
	}

	return nil
}

// GetBounds 获取节点位置
func (n *Node) GetBounds() (Rect, error) {
	var bounds Rect
	if n.handle == 0 {
		return bounds, errors.New("node handle is invalid")
	}

	proc, err := n.rpcClient.GetDLL().FindProc("getNodeNound")
	if err != nil {
		return bounds, fmt.Errorf("查找getNodeNound函数失败: %v", err)
	}

	ret, _, _ := proc.Call(
		n.handle,
		uintptr(unsafe.Pointer(&bounds.Left)),
		uintptr(unsafe.Pointer(&bounds.Top)),
		uintptr(unsafe.Pointer(&bounds.Right)),
		uintptr(unsafe.Pointer(&bounds.Bottom)),
	)

	if ret == 0 {
		return bounds, errors.New("获取节点位置失败")
	}

	return bounds, nil
}

// GetText 获取节点文本
func (n *Node) GetText() string {
	if n.handle == 0 {
		return ""
	}

	proc, err := n.rpcClient.GetDLL().FindProc("getNodeText")
	if err != nil {
		return ""
	}

	ptr, _, _ := proc.Call(n.handle)
	if ptr == 0 {
		return ""
	}

	// 安全地获取字符串
	b := make([]byte, 0)
	for i := 0; ; i++ {
		c := *(*byte)(unsafe.Pointer(uintptr(ptr) + uintptr(i)))
		if c == 0 {
			break
		}
		b = append(b, c)
	}
	text := string(b)

	// 释放内存
	if freeProc, err := n.rpcClient.GetDLL().FindProc("freeRpcPtr"); err == nil {
		freeProc.Call(ptr)
	}

	return text
}

// GetJSON 获取节点的JSON表示
func (n *Node) GetJSON() (string, error) {
	if n.handle == 0 {
		return "", errors.New("node handle is invalid")
	}

	proc, err := n.rpcClient.GetDLL().FindProc("getNodeJson")
	if err != nil {
		return "", fmt.Errorf("查找getNodeJson函数失败: %v", err)
	}

	ptr, _, _ := proc.Call(n.handle)
	if ptr == 0 {
		return "", errors.New("获取节点JSON失败")
	}

	// 安全地获取字符串
	b := make([]byte, 0)
	for i := 0; ; i++ {
		c := *(*byte)(unsafe.Pointer(uintptr(ptr) + uintptr(i)))
		if c == 0 {
			break
		}
		b = append(b, c)
	}
	jsonStr := string(b)

	// 释放内存
	if freeProc, err := n.rpcClient.GetDLL().FindProc("freeRpcPtr"); err == nil {
		freeProc.Call(ptr)
	}

	// 格式化JSON
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return jsonStr, nil // 如果解析失败，返回原始字符串
	}

	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return jsonStr, nil
	}

	return string(formatted), nil
}
