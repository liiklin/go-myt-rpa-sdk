package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"mytrpc/device"
	"mytrpc/node"
	"mytrpc/rpc"
)

func main() {
	// 打印命令行参数
	for i, arg := range os.Args[1:] {
		fmt.Printf("参数 %d: %s\n", i+1, arg)
	}

	// 解析命令行参数
	if len(os.Args) < 2 {
		fmt.Println("用法: mytrpc <host> [port]")
		fmt.Println("示例: mytrpc 192.168.1.181 11010")
		return
	}

	host := os.Args[1]
	port := 11010 // 默认端口

	if len(os.Args) > 2 {
		if p, err := strconv.Atoi(os.Args[2]); err == nil {
			port = p
		}
	}

	// 初始化RPC客户端
	client := rpc.NewClient()
	defer client.Close()

	// 连接设备
	if err := client.Connect(host, port); err != nil {
		log.Fatalf("连接设备失败: %v", err)
	}
	fmt.Printf("成功连接到设备 %s:%d\n", host, port)

	// 获取SDK版本
	version, err := client.GetSDKVersion()
	if err != nil {
		log.Fatalf("获取SDK版本失败: %v", err)
	}
	fmt.Printf("SDK 版本号: %s\n", version)

	// 检查连接状态
	connected, err := client.CheckConnectState()
	if err != nil {
		log.Fatalf("检查连接状态失败: %v", err)
	}
	if connected {
		fmt.Println("当前连接状态正常")
	} else {
		log.Fatal("当前连接断开")
	}

	// 初始化设备操作
	dev := device.NewDevice(client)

	// 设置RPA模式
	if err := dev.SetRPAMode(0); err != nil {
		log.Fatalf("设置RPA模式失败: %v", err)
	}
	fmt.Println("设置工作模式为 关闭无障碍")

	// // 长按操作示例
	// if err := dev.LongClick(1, 100, 100, 1.0); err != nil {
	// 	log.Printf("长按操作失败: %v", err)
	// }

	// // 截图并保存
	// if err := dev.SaveScreenshotToFile(device.ScreenshotOptions{
	// 	Quality: 90,
	// 	Region: image.Rectangle{
	// 		Min: image.Point{X: 0, Y: 0},
	// 		Max: image.Point{X: 0, Y: 0},
	// 	},
	// }, "screenshot_1.png"); err != nil {
	// 	log.Printf("保存截图失败: %v", err)
	// }

	// // 获取原始截图数据
	// screenshot, err := dev.GetScreenshot()
	// if err != nil {
	// 	log.Fatalf("获取截图失败: %v", err)
	// }
	// fmt.Printf("获取到截图数据: %d 字节\n", len(screenshot))

	// // 分区域截图示例
	// for i := 0; i < 3; i++ {
	// 	if err := dev.SaveScreenshotToFile(device.ScreenshotOptions{
	// 		Quality: 90,
	// 		Region: image.Rectangle{
	// 			Min: image.Point{X: i, Y: i},
	// 			Max: image.Point{X: 500 + i, Y: 500 + i},
	// 		},
	// 	}, fmt.Sprintf("screenshot_%d.png", i)); err != nil {
	// 		log.Printf("保存分区域截图失败: %v", err)
	// 	}
	// 	time.Sleep(500 * time.Millisecond)
	// }

	// // 执行命令示例
	// output, err := dev.ExecCmd("ls /data")
	// if err != nil {
	// 	log.Printf("执行命令失败: %v", err)
	// } else {
	// 	fmt.Printf("命令输出:\n%s\n", output)
	// }

	// 节点操作示例
	selector := node.NewSelector(client)
	selector.AddTextQuery("导入")
	node, err := selector.FindOne(200 * time.Millisecond)
	if err != nil {
		log.Printf("查找节点失败: %v", err)
	} else if node != nil {
		if err := node.Click(); err != nil {
			log.Printf("点击节点失败: %v", err)
		}
		json, _ := node.GetJSON()
		fmt.Printf("节点信息: %s\n", json)
	}

	// 应用管理测试
	if err := dev.OpenApp("com.blue.filemanager"); err != nil {
		log.Printf("打开文件管理器失败: %v", err)
	} else {
		fmt.Println("成功打开文件管理器")
	}

	time.Sleep(2 * time.Second)

	if err := dev.StopApp("com.blue.filemanager"); err != nil {
		log.Printf("关闭文件管理器失败: %v", err)
	} else {
		fmt.Println("成功关闭文件管理器")
	}

	// 文本输入测试
	if err := dev.SendText("中文测试!"); err != nil {
		log.Printf("发送文本失败: %v", err)
	} else {
		fmt.Println("成功发送文本")
	}

	if err := dev.ClearText(10); err != nil {
		log.Printf("清除文本失败: %v", err)
	} else {
		fmt.Println("成功清除文本")
	}

	// 其他测试用例可以根据需要继续添加...
}
