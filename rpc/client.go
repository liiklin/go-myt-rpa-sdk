package rpc

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

type Client struct {
	dll    *syscall.DLL
	handle uintptr
}

// NewClient 创建新的RPC客户端
func NewClient() *Client {
	return &Client{}
}

// getDLLPath 获取DLL文件路径
func getDLLPath() string {
	// 根据操作系统选择库文件
	var libName string
	switch runtime.GOOS {
	case "windows":
		libName = "libmytrpc.dll"
	case "linux":
		libName = "libmytrpc.so"
	case "darwin":
		libName = "libmytrpc.dylib"
	default:
		libName = "libmytrpc.dll"
	}

	// 尝试从当前目录的lib文件夹加载
	libPath := filepath.Join("lib", libName)
	fmt.Println("libPath:", libPath)
	if _, err := os.Stat(libPath); err == nil {
		return libPath
	}

	// 获取当前执行文件所在目录
	exePath, err := os.Executable()
	fmt.Println("exePath:", exePath)
	if err != nil {
		return "./lib/libmytrpc.dll"
	}
	exeDir := filepath.Dir(exePath)

	// 如果当前目录没有，则尝试从可执行文件目录加载
	return filepath.Join(exeDir, "lib", libName)
}

// Connect 连接到设备
func (c *Client) Connect(host string, port int) error {
	// 加载DLL
	var err error
	dllPath := getDLLPath()
	fmt.Println("正在加载DLL:", dllPath)
	c.dll, err = syscall.LoadDLL(dllPath)
	if err != nil {
		return fmt.Errorf("加载DLL失败(%s): %v", dllPath, err)
	}

	// 获取openDevice函数
	openDeviceProc, err := c.dll.FindProc("openDevice")
	if err != nil {
		return fmt.Errorf("查找openDevice函数失败: %v", err)
	}

	// 转换host为UTF-8字节数组
	hostBytes := []byte(host + "\x00") // 添加null终止符

	// 调用openDevice函数
	ret, _, err := openDeviceProc.Call(
		uintptr(unsafe.Pointer(&hostBytes[0])), // 传递字节数组指针
		uintptr(port),
		uintptr(10), // timeout
	)

	if ret == 0 {
		return fmt.Errorf("连接设备失败")
	}

	// 保存handle
	c.handle = ret

	// 等待连接建立
	fmt.Printf("正在连接设备 %s:%d...\n", host, port)
	time.Sleep(1 * time.Second)

	return nil
}

// GetSDKVersion 获取SDK版本
func (c *Client) GetSDKVersion() (string, error) {
	proc, err := c.dll.FindProc("getVersion")
	if err != nil {
		return "", fmt.Errorf("查找getVersion函数失败: %v", err)
	}

	ret, _, err := proc.Call()
	if ret == 0 {
		return "", fmt.Errorf("获取版本失败")
	}

	return fmt.Sprintf("%d", ret), nil
}

// CheckConnectState 检查连接状态
func (c *Client) CheckConnectState() (bool, error) {
	proc, err := c.dll.FindProc("checkLive")
	if err != nil {
		return false, fmt.Errorf("查找checkLive函数失败: %v", err)
	}

	ret, _, err := proc.Call(uintptr(c.handle))
	return ret != 0, nil
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	if c.handle != 0 {
		if closeProc, err := c.dll.FindProc("closeDevice"); err == nil {
			closeProc.Call(uintptr(c.handle))
		}
	}
	if c.dll != nil {
		return c.dll.Release()
	}
	return nil
}

// GetDLL 获取DLL实例
func (c *Client) GetDLL() *syscall.DLL {
	return c.dll
}

// GetHandle 获取设备句柄
func (c *Client) GetHandle() uintptr {
	return c.handle
}
