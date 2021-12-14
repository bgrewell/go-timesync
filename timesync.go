package timesync

import "time"

var (
	clock Clock
)

func init() {
	clock = &SimpleClock{}
	clock.SetInterval(5 * time.Second)
}

func SyncOnce(host string, port int) error {
	err := clock.EnableSync(host, port)
	clock.DisableSync()
	return err
}

func EnableSync(host string, port int) error {
	return clock.EnableSync(host, port)
}

func DisableSync() {
	clock.DisableSync()
}

func Now() time.Time {
	return clock.Now()
}

func Offset() time.Duration {
	return clock.Offset()
}

func Delay() time.Duration {
	return clock.Delay()
}
