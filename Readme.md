# 架构：核心-适配器模式

这个架构只包含两个主要层级：

1. **核心业务层 (`core`)：** 负责数据处理、日志关联、策略执行。它是平台无关的。
2. **监控适配器层 (`monitor`)：** 负责与操作系统（Ubuntu/Linux）底层 API 交互。**它是平台特有的，也是未来移植时唯一需要修改的部分。**

## 流程概览

1. `monitor` 包内的 Goroutine 持续监听 Linux 事件（`udev`, `inotify`）。
2. 事件发生后，适配器将原始数据**封装**成标准的 `Event` 结构体。
3. 适配器通过 `channel `将 `Event` 发送到 `core` 包。
4. `core` 包（关联器）从 Channel 读取事件，进行 PID 关联、日志记录。

# 目录结构

```text
usbMonitor/
├── cmd/
│   └── monitor/
│       └── main.go       # 程序入口，启动核心引擎
├── internal/
│   ├── core/             # 核心业务逻辑 (平台无关)
│   │   ├── engine.go     # 核心引擎，启动所有监控器
│   │   └── correlator.go # 日志关联器
│   ├── monitor/          # 监控适配器层
│   │   ├── monitor.go    # 定义所有监控器接口 (e.g., DeviceMonitor Interface)
│   │   ├── linux/        # 仅限 Linux 的实现代码 (使用 Go Build Tag)
│   │   │   ├── device.go     # 实现 DeviceMonitor (使用 udev/sysfs)
│   │   │   └── fs.go         # 实现 FSMonitor (使用 fsnotify/inotify)
│   │   └── windows/      # 未来 Windows 的实现代码 (使用 Go Build Tag)
│   │       ├── device.go     # 实现 DeviceMonitor (使用 WMI/PnP)
│   │       └── fs.go         # 实现 FSMonitor (使用 ReadDirectoryChangesW)
│   └── config/           # 配置管理
├── pkg/                  # 公共包 (平台无关的工具和结构体)
│   ├── event/            # 标准化数据结构 (MonitorEvent)
│   └── logging/          # 通用日志记录工具
├── go.mod
├── go.sum
└── README.md
```

# 功能

* [X] 设备插拔事件监控。实时记录 USB 设备在何时连接/断开系统
* [X] 设备信息采集。记录设备的供应商 ID (VID)、产品 ID (PID)、序列号、设备类型（U盘、键盘、鼠标等）
* [X] 文件系统操作监控。监控在可移动存储设备（如 U盘）上发生的文件读写、创建、删除操作。这是最核心的“流量”表现
* [ ] 记录哪个进程 （例如资源管理器、复制工具）对 USB 设备上的文件进行了操作：使用fanotify（Linux，root模式下）
* [ ] 安全功能该如何体现呢？

# 运行命令

`go run cmd/monitor/main.go`


test github ssh
