package utils

import "time"

func Millisecond(ms int64) time.Duration {
	return time.Duration(ms) * time.Millisecond
}
