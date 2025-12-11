package linux_monitor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/usbMonitor/pkg/event"
)

type DeviceMonitor struct {
	stopChan chan struct{}
}

func NewDeviceMonitor() *DeviceMonitor {
	return &DeviceMonitor{
		stopChan: make(chan struct{}),
	}
}

// 获取当前 USB 设备列表
func getUSBDevices() (map[string]bool, error) {
	devices := make(map[string]bool)
	files, err := os.ReadDir("/sys/bus/usb/devices")
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		name := f.Name()

		// 过滤掉接口（如2-1:1.0）和总线（如usb1）等非设备项
		if !strings.Contains(name, ":") && !strings.HasPrefix(name, "usb") {
			devices[name] = true
		}
	}
	return devices, nil
}

// 读取指定路径下文件的内容并返回字符串
func readFile(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return "N/A"
	}
	return strings.TrimSpace(string(content))
}

func readDeviceAttributes(deviceID string) map[string]string {
	details := make(map[string]string)

	// 构建设备目录路径
	devicePath := filepath.Join("/sys/bus/usb/devices", deviceID)

	// 读取VID和PID
	details["vid"] = readFile(filepath.Join(devicePath, "idVendor"))
	details["pid"] = readFile(filepath.Join(devicePath, "idProduct"))

	// 读取设备序列号
	details["serial"] = readFile(filepath.Join(devicePath, "serial"))

	// 读取设备类别
	// 常见的类别代码：08（存储设备）；03（HID/鼠标键盘）
	class_code := readFile(filepath.Join(devicePath, "bDeviceClass"))
	details["class_code"] = class_code
	switch class_code {
	case "08":
		details["class_name"] = "Mass Storage"
	case "03":
		details["class_name"] = "Human Interface Device (HID)"
	default:
		details["class_name"] = "Other"
	}

	// 获取设备描述符中的产品名称
	details["product_name"] = readFile(filepath.Join(devicePath, "product"))

	return details
}

func (d *DeviceMonitor) Start() (<-chan event.MonitorEvent, error) {
	eventChan := make(chan event.MonitorEvent)

	//获取当前设备列表
	currentDevices, _ := getUSBDevices()

	go func() {
		ticker := time.NewTicker(1 * time.Second) // 每1秒扫描一次
		defer ticker.Stop()
		defer close(eventChan)

		for {
			select {
			case <-d.stopChan:
				return
			case <-ticker.C:
				// 获取后续设备列表
				newDevices, err := getUSBDevices()
				if err != nil {
					continue
				}

				// 检查新增设备
				for devID := range newDevices {
					if !currentDevices[devID] {

						details := readDeviceAttributes(devID)

						eventChan <- event.MonitorEvent{
							Timestamp: time.Now(),
							Type:      "DEVICE_ADD",
							Source:    "DEVICE_MONITOR (Linux)",
							Message:   fmt.Sprintf("Device Added: %s (%s:%s)", details["product_name"], details["vid"], details["pid"]),
							Details:   details,
						}
					}
				}

				// 检查移除设备
				for devID := range currentDevices {
					if !newDevices[devID] {
						eventChan <- event.MonitorEvent{
							Timestamp: time.Now(),
							Type:      "DEVICE_REMOVE",
							Source:    "DEVICE_MONITOR (Linux)",
							Message:   "USB Device Removed: " + devID,
							Details:   map[string]string{"id": devID},
						}
					}
				}

				// 更新状态
				currentDevices = newDevices
			}
		}
	}()
	return eventChan, nil
}

func (d *DeviceMonitor) Stop() {
	close(d.stopChan)
}
