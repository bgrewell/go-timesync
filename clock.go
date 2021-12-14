package timesync

import "time"

type Clock interface {
	Now() time.Time
	EnableSync(host string, port int) error
	DisableSync()
	EnableSystemTimeAdjustment()
	DisableSystemTimeAdjustment()
	SetInterval(interval time.Duration)
	GetInterval() time.Duration
	LastSync() time.Time
	Offset() time.Duration
	Delay() time.Duration
	EnableServer(listenAddr string, port int) error
	DisableServer()
}
