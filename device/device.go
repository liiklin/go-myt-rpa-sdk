package device

import (
	"errors"
	"fmt"
	"image"
	"mytrpc/rpc"
	"os"
	"syscall"
	"time"
	"unsafe"
)

type Device struct {
	client *rpc.Client
}

func NewDevice(client *rpc.Client) *Device {
	return &Device{
		client: client,
	}
}

type ScreenshotOptions struct {
	Quality int
	Region  image.Rectangle
}

type KeyCode int

const (
	KeyCodeVolumeUp KeyCode = iota + 24
	KeyCodeVolumeDown
	KeyCodePower
	KeyCodeHome
	KeyCodeBack
)

type SwipeOptions struct {
	StartX, StartY int
	EndX, EndY     int
	Duration       time.Duration
}

func (d *Device) SetRPAMode(mode int) error {
	proc, err := d.client.GetDLL().FindProc("useNewNodeMode")
	if err != nil {
		return fmt.Errorf("查找useNewNodeMode函数失败: %v", err)
	}

	ret, _, err := proc.Call(
		uintptr(d.client.GetHandle()),
		uintptr(mode),
	)

	if ret == 0 {
		return fmt.Errorf("设置RPA模式失败")
	}

	return nil
}

func (d *Device) TakeScreenshot(opts ScreenshotOptions) ([]byte, error) {
	proc, err := d.client.GetDLL().FindProc("takeCaptrueCompress")
	if err != nil {
		return nil, fmt.Errorf("查找takeCaptrueCompress函数失败: %v", err)
	}

	var dataLen int32
	ret, _, err := proc.Call(
		uintptr(d.client.GetHandle()),
		uintptr(0), // type: 0 for PNG
		uintptr(opts.Quality),
		uintptr(unsafe.Pointer(&dataLen)),
	)

	if ret == 0 {
		return nil, errors.New("截图失败")
	}

	// 安全地复制数据
	data := make([]byte, dataLen)
	if dataLen > 0 {
		dataPtr := (*[1 << 30]byte)(unsafe.Pointer(ret))
		copy(data, dataPtr[:dataLen])
	}

	// 释放DLL分配的内存
	if freeProc, err := d.client.GetDLL().FindProc("freeRpcPtr"); err == nil {
		freeProc.Call(ret)
	}

	return data, nil
}

func (d *Device) GetScreenshot() ([]byte, error) {
	return d.TakeScreenshot(ScreenshotOptions{
		Quality: 90,
		Region: image.Rectangle{
			Min: image.Point{X: 0, Y: 0},
			Max: image.Point{X: 0, Y: 0},
		},
	})
}

func (d *Device) KeyPress(code KeyCode) error {
	proc, err := d.client.GetDLL().FindProc("keyPress")
	if err != nil {
		return fmt.Errorf("查找keyPress函数失败: %v", err)
	}

	ret, _, err := proc.Call(
		uintptr(d.client.GetHandle()),
		uintptr(code),
	)
	if ret == 0 {
		return errors.New("按键操作失败")
	}

	return nil
}

func (d *Device) Swipe(opts SwipeOptions) error {
	proc, err := d.client.GetDLL().FindProc("swipe")
	if err != nil {
		return fmt.Errorf("查找swipe函数失败: %v", err)
	}

	ret, _, err := proc.Call(
		uintptr(d.client.GetHandle()),
		uintptr(1),
		uintptr(opts.StartX),
		uintptr(opts.StartY),
		uintptr(opts.EndX),
		uintptr(opts.EndY),
		uintptr(opts.Duration.Milliseconds()),
		uintptr(0),
	)

	if ret == 0 {
		return errors.New("滑动操作失败")
	}

	return nil
}

// LongClick 长按操作
func (d *Device) LongClick(fingerID int, x, y int, duration float64) error {
	// 按下
	proc, err := d.client.GetDLL().FindProc("touchDown")
	if err != nil {
		return fmt.Errorf("查找touchDown函数失败: %v", err)
	}

	ret, _, err := proc.Call(
		uintptr(d.client.GetHandle()),
		uintptr(fingerID),
		uintptr(x),
		uintptr(y),
	)
	if ret == 0 {
		return errors.New("按下操作失败")
	}

	// 等待指定时间
	time.Sleep(time.Duration(duration * float64(time.Second)))

	// 抬起
	proc, err = d.client.GetDLL().FindProc("touchUp")
	if err != nil {
		return fmt.Errorf("查找touchUp函数失败: %v", err)
	}

	ret, _, err = proc.Call(
		uintptr(d.client.GetHandle()),
		uintptr(fingerID),
		uintptr(x),
		uintptr(y),
	)
	if ret == 0 {
		return errors.New("抬起操作失败")
	}

	return nil
}

// SaveScreenshotToFile 保存截图到文件
func (d *Device) SaveScreenshotToFile(opts ScreenshotOptions, filePath string) error {
	// 尝试不同的函数名
	var proc *syscall.Proc
	var err error

	for _, funcName := range []string{"screentshotEx", "ScreentshotEx", "screentShotEx"} {
		if proc, err = d.client.GetDLL().FindProc(funcName); err == nil {
			break
		}
	}

	if err != nil {
		// 如果都失败了，尝试使用TakeScreenshot并手动保存
		data, err := d.TakeScreenshot(opts)
		if err != nil {
			return fmt.Errorf("截图失败: %v", err)
		}

		// 创建文件并写入数据
		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("创建文件失败: %v", err)
		}
		defer file.Close()

		if _, err := file.Write(data); err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}

		return nil
	}

	ret, _, err := proc.Call(
		uintptr(d.client.GetHandle()),
		uintptr(opts.Region.Min.X),
		uintptr(opts.Region.Min.Y),
		uintptr(opts.Region.Max.X),
		uintptr(opts.Region.Max.Y),
		uintptr(1), // type: 1 for JPG
		uintptr(opts.Quality),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(filePath))),
	)

	if ret == 0 {
		return errors.New("保存截图失败")
	}

	return nil
}

// ExecCmd 执行命令
func (d *Device) ExecCmd(cmd string) (string, error) {
	proc, err := d.client.GetDLL().FindProc("execCmd")
	if err != nil {
		return "", fmt.Errorf("查找execCmd函数失败: %v", err)
	}

	cmdBytes := []byte(cmd + "\x00")
	ret, _, err := proc.Call(
		uintptr(d.client.GetHandle()),
		uintptr(1), // sync mode
		uintptr(unsafe.Pointer(&cmdBytes[0])),
	)

	if ret == 0 {
		return "", errors.New("执行命令失败")
	}

	// 使用 unsafe.String 替代 syscall.UTF16PtrToString
	ptr := (*uint16)(unsafe.Pointer(ret))
	n := 0
	for *(*uint16)(unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + uintptr(n)*2)) != 0 {
		n++
	}
	return unsafe.String((*byte)(unsafe.Pointer(ptr)), n*2), nil
}

// OpenApp 打开应用
func (d *Device) OpenApp(packageName string) error {
	proc, err := d.client.GetDLL().FindProc("openApp")
	if err != nil {
		return fmt.Errorf("查找openApp函数失败: %v", err)
	}

	pkgBytes := []byte(packageName + "\x00")
	ret, _, err := proc.Call(
		uintptr(d.client.GetHandle()),
		uintptr(unsafe.Pointer(&pkgBytes[0])),
	)

	if ret == 0 {
		return fmt.Errorf("打开应用 %s 失败", packageName)
	}

	return nil
}

// StopApp 关闭应用
func (d *Device) StopApp(packageName string) error {
	proc, err := d.client.GetDLL().FindProc("stopApp")
	if err != nil {
		return fmt.Errorf("查找stopApp函数失败: %v", err)
	}

	pkgBytes := []byte(packageName + "\x00")
	ret, _, err := proc.Call(
		uintptr(d.client.GetHandle()),
		uintptr(unsafe.Pointer(&pkgBytes[0])),
	)

	if ret == 0 {
		return fmt.Errorf("关闭应用 %s 失败", packageName)
	}

	return nil
}

// SendText 输入文本
func (d *Device) SendText(text string) error {
	proc, err := d.client.GetDLL().FindProc("sendText")
	if err != nil {
		return fmt.Errorf("查找sendText函数失败: %v", err)
	}

	textBytes := []byte(text + "\x00")
	ret, _, err := proc.Call(
		uintptr(d.client.GetHandle()),
		uintptr(unsafe.Pointer(&textBytes[0])),
	)

	if ret == 0 {
		return errors.New("发送文本失败")
	}

	return nil
}

// ClearText 清除文本
func (d *Device) ClearText(count int) error {
	// 使用 sendText 发送退格键
	proc, err := d.client.GetDLL().FindProc("keyPress")
	if err != nil {
		return fmt.Errorf("查找keyPress函数失败: %v", err)
	}

	// 发送count个退格键
	for i := 0; i < count; i++ {
		ret, _, _ := proc.Call(
			uintptr(d.client.GetHandle()),
			uintptr(67),
		)

		if ret == 0 {
			return errors.New("清除文本失败")
		}
	}

	return nil
}

// DumpNodeXml 导出节点XML信息
func (d *Device) DumpNodeXml(dumpAll bool) (string, error) {
	proc, err := d.client.GetDLL().FindProc("dumpNodeXml")
	if err != nil {
		return "", fmt.Errorf("查找dumpNodeXml函数失败: %v", err)
	}

	var mode uintptr
	if dumpAll {
		mode = 1
	}

	ret, _, _ := proc.Call(d.client.GetHandle(), mode)
	if ret == 0 {
		return "", fmt.Errorf("导出节点XML失败")
	}

	// 转换返回的字符串
	ptr := (*byte)(unsafe.Pointer(ret))
	n := 0
	for *(*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + uintptr(n))) != 0 {
		n++
	}
	bytes := make([]byte, n)
	copy(bytes, unsafe.Slice(ptr, n))

	// 释放内存
	if freeProc, err := d.client.GetDLL().FindProc("freeRpcPtr"); err == nil {
		freeProc.Call(ret)
	}

	return string(bytes), nil
}

// DumpNodeXmlEx 导出节点XML信息（带工作模式和超时参数）
func (d *Device) DumpNodeXmlEx(workMode bool, timeout int) (string, error) {
	proc, err := d.client.GetDLL().FindProc("dumpNodeXmlEx")
	if err != nil {
		return "", fmt.Errorf("查找dumpNodeXmlEx函数失败: %v", err)
	}

	var mode uintptr
	if workMode {
		mode = 1
	}

	ret, _, _ := proc.Call(d.client.GetHandle(), mode, uintptr(timeout))
	if ret == 0 {
		return "", fmt.Errorf("导出节点XML失败")
	}

	// 转换返回的字符串
	ptr := (*byte)(unsafe.Pointer(ret))
	n := 0
	for *(*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + uintptr(n))) != 0 {
		n++
	}
	bytes := make([]byte, n)
	copy(bytes, unsafe.Slice(ptr, n))

	// 释放内存
	if freeProc, err := d.client.GetDLL().FindProc("freeRpcPtr"); err == nil {
		freeProc.Call(ret)
	}

	return string(bytes), nil
}

func (d *Device) TouchDown(x, y int, fingerID int) error {
	proc, err := d.client.GetDLL().FindProc("touchDown")
	if err != nil {
		return fmt.Errorf("查找touchDown函数失败: %v", err)
	}
	ret, _, _ := proc.Call(
		uintptr(d.client.GetHandle()),
		uintptr(fingerID),
		uintptr(x),
		uintptr(y),
	)
	if ret == 0 {
		return errors.New("触摸按下失败")
	}
	return nil
}

func (d *Device) TouchUp(x, y int, fingerID int) error {
	proc, err := d.client.GetDLL().FindProc("touchUp")
	if err != nil {
		return fmt.Errorf("查找touchUp函数失败: %v", err)
	}

	ret, _, _ := proc.Call(
		uintptr(d.client.GetHandle()),
		uintptr(fingerID),
		uintptr(x),
		uintptr(y),
	)
	if ret == 0 {
		return errors.New("触摸抬起失败")
	}
	return nil
}
