package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/Hara602/usbMonitor/internal/core"
	linux_monitor "github.com/Hara602/usbMonitor/internal/monitor/linux"
	"github.com/Hara602/usbMonitor/pkg/logging"
)

func main() {
	// 初始化日志系统
	// 设置为开发模式，日志级别为 Debug，方便排错
	logging.InitLogger("development", "debug")
	defer logging.CloseLogger() // 确保在 main 函数退出时刷新日志缓冲区

	// 确定监控根目录
	// 尝试获取当前用户，构建 /media/username 目录
	var watchPath string
	currentUser, err := user.Current()

	// 如果是 root 运行，可能需要监控整个 /media
	if err == nil && currentUser.Username != "root" {
		watchPath = filepath.Join("/media", currentUser.Username)
	} else {
		watchPath = "/media"
	}

	// 检查目录是否存在，不存在则回退到 /media
	// os.IsNotExist 用于判断一个错误是否表示“文件或目录不存在”
	if _, err := os.Stat(watchPath); os.IsNotExist(err) {
		watchPath = "/media"
	}

	fmt.Printf("============================================\n")
	fmt.Printf("🛡️ USB Monitor Started In Linux\n")
	fmt.Printf("📂 Watching Mount Point: %s\n", watchPath)
	fmt.Printf("============================================\n")

	// 2. 初始化核心引擎
	engine := core.NewEngine()

	// 3. 添加 Linux 设备监控适配器 (监控 /sys 变化)
	deviceMon := linux_monitor.NewDeviceMonitor()
	engine.AddMonitor(deviceMon)

	// 4. 添加 Linux 文件监控适配器 (真实路径)
	fsMon := linux_monitor.NewFSMonitor(watchPath)
	engine.AddMonitor(fsMon)

	// 5. 运行
	engine.Run()
}
