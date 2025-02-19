package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"mytrpc/device"
	"mytrpc/rpc"

	"golang.org/x/exp/rand"
)

// 设备信息结构体
type DeviceInfo struct {
	Host   string
	Port   int
	Device *device.Device
	Wg     sync.WaitGroup
	Ctx    context.Context
	Cancel context.CancelFunc
}

var (
	devices     = make(map[string]*DeviceInfo) // 设备映射表 key格式: IP_Port
	devicesLock sync.RWMutex
)

// 执行回复操作
func doReply(dev *device.Device, i int, wg *sync.WaitGroup, sendText string) {
	defer wg.Done()

	log.Println("第", i+1, "次开始")
	// 按下键盘esc
	dev.KeyPress(111)
	time.Sleep(1000 * time.Millisecond) // 增加等待时间

	// 点击More按钮
	dev.TouchDown(660, 1200, 1)
	time.Sleep(1500 * time.Millisecond) // 增加触摸时间
	dev.TouchUp(660, 1200, 1)
	time.Sleep(1500 * time.Millisecond) // 增加等待时间确保界面切换

	// 点击Comment按钮
	dev.TouchDown(200, 1000, 2)
	time.Sleep(1500 * time.Millisecond) // 增加触摸时间
	dev.TouchUp(200, 1000, 2)
	time.Sleep(1500 * time.Millisecond) // 增加等待时间确保输入框就绪

	// 输入评论（带重试机制）
	for retry := 0; retry < 3; retry++ {
		dev.ClearText(1000)
		time.Sleep(800 * time.Millisecond)
		if err := dev.SendText(sendText); err == nil {
			log.Println("发送文字成功!")
			time.Sleep(500 * time.Millisecond)
			// 点击发送
			dev.KeyPress(66)
			break
		} else {
			log.Printf("发送文字失败(第%d次重试): %v", retry+1, err)
			time.Sleep(1 * time.Second)
		}
	}

	time.Sleep(300 * time.Millisecond) // 增加结束等待
	log.Println("第", i+1, "次结束")
}

// 初始化设备连接
func initDevice(host string, port int) (*DeviceInfo, error) {
	deviceID := fmt.Sprintf("%s_%d", host, port)

	// 检查设备是否已存在
	devicesLock.RLock()
	if dev, exists := devices[deviceID]; exists {
		devicesLock.RUnlock()
		return dev, nil
	}
	devicesLock.RUnlock()

	// 创建新设备连接
	client := rpc.NewClient()
	if err := client.Connect(host, port); err != nil {
		return nil, fmt.Errorf("连接设备失败: %v", err)
	}

	// 检查连接状态
	if connected, err := client.CheckConnectState(); err != nil || !connected {
		return nil, fmt.Errorf("设备连接验证失败")
	}

	// 创建设备上下文
	ctx, cancel := context.WithCancel(context.Background())

	dev := &DeviceInfo{
		Host:   host,
		Port:   port,
		Device: device.NewDevice(client),
		Ctx:    ctx,
		Cancel: cancel,
	}

	// 保存到设备映射表
	devicesLock.Lock()
	devices[deviceID] = dev
	devicesLock.Unlock()

	return dev, nil
}

// 修改设备任务执行方式为顺序执行
func startDeviceTask(host string, port int) {
	deviceID := fmt.Sprintf("%s_%d", host, port)

	dev, err := initDevice(host, port)
	if err != nil {
		log.Printf("[%s] 初始化失败: %v", deviceID, err)
		return
	}

	log.Printf("[%s] 开始执行任务", deviceID)

	// 修改为顺序执行，每次循环等待完成
	for i := 0; i < 100; i++ {
		select {
		case <-dev.Ctx.Done():
			log.Printf("[%s] 任务已终止", deviceID)
			return
		default:
			// 移除了goroutine，直接顺序执行
			dev.Wg.Add(1)
			sendText := strconv.Itoa(rand.Intn(1000000))
			doReply(dev.Device, i, &dev.Wg, sendText)
		}
	}
	dev.Wg.Wait()
}

func main() {
	if len(os.Args) < 3 {
		log.Println("用法: mytrpc <host> <port1> [port2] ...")
		log.Println("示例: mytrpc 192.168.1.181 7101 7102")
		return
	}

	host := os.Args[1]
	ports := make([]int, 0, len(os.Args)-2)

	// 解析端口参数
	for _, arg := range os.Args[2:] {
		port, err := strconv.Atoi(arg)
		if err != nil {
			log.Printf("无效的端口号: %s", arg)
			continue
		}
		ports = append(ports, port)
	}

	if len(ports) == 0 {
		log.Println("至少需要指定一个端口号")
		return
	}

	// 为每个端口启动独立任务
	var mainWg sync.WaitGroup
	for _, port := range ports {
		mainWg.Add(1)
		go func(p int) {
			defer mainWg.Done()
			startDeviceTask(host, p)
		}(port)
	}

	// 等待所有任务完成
	mainWg.Wait()

	// 清理资源
	devicesLock.RLock()
	defer devicesLock.RUnlock()
	for _, dev := range devices {
		dev.Cancel()
	}
}
