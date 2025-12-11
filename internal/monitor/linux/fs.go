package linux_monitor

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/usbMonitor/pkg/event"

	"github.com/fsnotify/fsnotify"
)

type FSMonitor struct {
	watchRoot string // 通常是 /media 或 /media/username
	watcher   *fsnotify.Watcher
}

func NewFSMonitor(rootPath string) *FSMonitor {
	return &FSMonitor{watchRoot: rootPath}
}

// Helper: 递归添加目录及其子目录到监控列表
func (f *FSMonitor) addRecursive(path string) {
	err := filepath.Walk(path, func(walkPath string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			err = f.watcher.Add(walkPath)
			if err != nil {
				log.Printf("Failed to watch directory: %s, error: %v", walkPath, err)
			} else {
				// log.Printf("Started watching: %s", walkPath) // 调试用，可以注释掉
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("Error walking path %s: %v", path, err)
	}
}

func (f *FSMonitor) Start() (<-chan event.MonitorEvent, error) {
	var err error
	f.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// 1. 初始扫描：添加根目录及其当前所有子目录
	// 这样如果你启动程序时 U 盘已经插着，也能被监控到
	f.addRecursive(f.watchRoot)

	eventChan := make(chan event.MonitorEvent)

	go func() {
		defer close(eventChan)
		defer f.watcher.Close()

		for {
			select {
			case fsEvent, ok := <-f.watcher.Events:
				if !ok {
					return
				}

				// 忽略一些噪音事件 (Chmod)
				if fsEvent.Op&fsnotify.Chmod == fsnotify.Chmod {
					continue
				}

				// 2. 关键逻辑：如果是“创建”事件，且创建的是“目录”，则将其加入监控
				// 这对应了 U 盘刚刚挂载的情况
				if fsEvent.Op&fsnotify.Create == fsnotify.Create {

					// 等待挂载完成。系统创建目录到挂载文件系统通常需要几十到几百毫秒。
					// 如果不等待，可能会遇到 "permission denied" 或者扫描空目录。
					// TODO: 需要一个更优雅的方式，而不是简单的 sleep。
					time.Sleep(1 * time.Second)

					fi, err := os.Stat(fsEvent.Name)
					if err == nil && fi.IsDir() {
						log.Printf("[FS] New mount/directory detected: %s", fsEvent.Name)
						f.addRecursive(fsEvent.Name)
					}
				}

				// 转换事件类型
				eventType := "UNKNOWN"
				if fsEvent.Op&fsnotify.Write == fsnotify.Write {
					eventType = "FILE_WRITE"
				} else if fsEvent.Op&fsnotify.Create == fsnotify.Create {
					eventType = "FILE_CREATE"
				} else if fsEvent.Op&fsnotify.Remove == fsnotify.Remove {
					eventType = "FILE_DELETE"
				} else if fsEvent.Op&fsnotify.Rename == fsnotify.Rename {
					eventType = "FILE_RENAME"
				}

				if eventType != "UNKNOWN" {
					eventChan <- event.MonitorEvent{
						Timestamp: time.Now(),
						Type:      eventType,
						Source:    "FS_MONITOR",
						Message:   "File activity: " + filepath.Base(fsEvent.Name),
						Details: map[string]string{
							"FilePath": fsEvent.Name,
							"Action":   fsEvent.Op.String(),
						},
					}
				}

			case err, ok := <-f.watcher.Errors:
				if !ok {
					return
				}
				log.Println("FS Monitor Error:", err)
			}
		}
	}()

	return eventChan, nil
}

func (f *FSMonitor) Stop() {
	f.watcher.Close()
}
