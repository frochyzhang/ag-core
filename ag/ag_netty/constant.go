package ag_netty

import "time"

var TimeoutUnit = time.Millisecond

func ToTimeoutDuration(timeout int) time.Duration {
	return time.Duration(timeout) * TimeoutUnit
}
