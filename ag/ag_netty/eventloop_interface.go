package ag_netty

import "time"

// EventLooper 事件循环接口
type EventLooper interface {
	Post(task func())
	Schedule(delay time.Duration, task func())
	Shutdown()
	IsShutdown() bool
}
