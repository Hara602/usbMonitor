package event

import "time"

// MonitorEvent 定义标准化的事件结构
type MonitorEvent struct {
	Timestamp time.Time
	Type      string            // "DEVICE_ADD", "DEVICE_REMOVE", "FILE_WRITE", "FILE_CREATE"
	Source    string            // "DEVICE_MONITOR", "FS_MONITOR"
	Message   string            // 人类可读的消息
	Details   map[string]string //// 存储 VID, PID, FilePath, ProcessName 等
}
