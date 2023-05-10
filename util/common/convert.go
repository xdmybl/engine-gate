package common

func Int64ToUint32(i int64) uint32 {
	if i < -1 {
		return 0
	} else if i == -1 || i == 0 || i > 4294967295 {
		return 4294967295 // 2^32 - 1
	} else {
		return uint32(i)
	}
}
