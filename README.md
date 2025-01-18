# MYT RPC SDK Go 版本

这是一个用于与MYT设备进行RPC通信的Go语言SDK。

## 功能特性

- 设备连接管理
- 设备操作（截图、按键、滑动等）
- 节点选择与操作
- 命令执行
- 超时控制

## 安装

1. 确保已安装Go 1.20或更高版本
2. 克隆仓库：
   ```bash
   git clone https://github.com/your-repo/myt-rpc-go.git
   ```
3. 进入项目目录：
   ```bash
   cd myt-rpc-go
   ```
4. 安装依赖：
   ```bash
   go mod tidy
   ```

## 使用示例

```go
package main

import (
	"log"
	"mytrpc/rpc"
)

func main() {
	client := rpc.NewClient()
	defer client.Close()

	if err := client.Connect("192.168.1.100", 11010); err != nil {
		log.Fatal(err)
	}

	version, err := client.GetSDKVersion()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("SDK Version:", version)
}
```

## API 文档

完整的API文档请参考 [API文档](docs/api.md)

## 贡献

欢迎提交issue和pull request

## 许可证

MIT License