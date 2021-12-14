package timesync

func CalcOffset(t1, t2, t3, t4 int64) int64 {
	return ((t2 - t1) + (t3 - t4)) / 2
}

func CalcDelay(t1, t2, t3, t4 int64) int64 {
	return (t4 - t1) - (t3 - t2)
}
