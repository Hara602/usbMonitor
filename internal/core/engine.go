package core

import (
	"fmt"
	"sync"

	"github.com/Hara602/usbMonitor/internal/monitor"
	"github.com/Hara602/usbMonitor/pkg/event"
)

type Engine struct {
	monitors []monitor.MonitorInterface
}

func NewEngine() *Engine {
	return &Engine{
		monitors: []monitor.MonitorInterface{},
	}
}

func (e *Engine) AddMonitor(m monitor.MonitorInterface) {
	e.monitors = append(e.monitors, m)
}

func (e *Engine) Run() {
	fmt.Println("[Core] Starting Security Engine...")

	// 统一的事件汇聚通道
	aggregator := make(chan event.MonitorEvent)
	var wg sync.WaitGroup

	// 启动所有添加的监控器
	for _, m := range e.monitors {
		// 监控器启动，将监控事件以通道形式返回给ch
		ch, err := m.Start()
		if err != nil {
			fmt.Printf("[Error] Failed to start monitor: %v\n", err)
			continue
		}

		wg.Add(1)
		// 启动协程将各个监控器的事件转发到总通道
		go func(c <-chan event.MonitorEvent) {
			defer wg.Done()
			for evt := range c {
				aggregator <- evt
			}
		}(ch)
	}

	// 核心业务逻辑：处理汇聚后的事件
	// 这里可以加入 策略引擎、日志关联 等逻辑
	go func() {
		for evt := range aggregator {

			// --- 日志输出 ---
			fmt.Println("------------------------------------------------")
			// 使用增强后的信息输出
			fmt.Printf("🔴 [%s] EVENT: %s\n", evt.Timestamp.Format("15:04:05"), evt.Type)
			fmt.Printf("   Source: %s\n", evt.Source)
			fmt.Printf("   Message: %s\n", evt.Message)

			// 打印进程关联信息
			if pid, ok := evt.Details["ProcessID"]; ok {
				fmt.Printf("   >> Process: %s (PID %s) by User: %s\n",
					evt.Details["ProcessName"], pid, evt.Details["ProcessUser"])
			}

			if path, ok := evt.Details["FilePath"]; ok {
				fmt.Printf("   >> File Affected: %s\n", path)
			}
			fmt.Println("------------------------------------------------")
		}
	}()

	// 阻塞主线程
	wg.Wait()
}
