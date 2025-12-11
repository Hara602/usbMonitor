package monitor

import "github.com/Hara602/usbMonitor/pkg/event"

// MonitorInterface 是所有监控器必须实现的接口
// 核心层不需要知道底层是 Linux 还是 Windows
type MonitorInterface interface {
	Start() (<-chan event.MonitorEvent, error)
	Stop()
}
