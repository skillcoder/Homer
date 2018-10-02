package helper

/* vim: set ts=2 sw=2 sts=2 ff=unix ft=go noet: */

import (
	"sort"
)

// GetPrevTimestampLoop DEPRECATED return previous unix_timestamp rounded by secPeriod, use GetPrevTimestamp instead
func GetPrevTimestampLoop(timestamp int64, secPeriod int) (ts int64) {
	ts = timestamp
	for {
		if ts%int64(secPeriod) == 0 {
			return ts
		}
		ts--
	}
}

// GetPrevTimestamp return previous unix_timestamp rounded by secPeriod
func GetPrevTimestamp(timestamp int64, secPeriod int) (ts int64) {
	return timestamp - (timestamp % int64(secPeriod))
}

// GetNextMinute return next minute start time in unix_timestamp formate
func GetNextMinute(timestamp int64) (nextMinute int64) {
	prev := timestamp - (timestamp % 60)
	nextMinute = prev + 60
	return nextMinute
}

// InsertSortedInt64 insert into slice []int64 by order
// https://gist.github.com/zhum/57cb45d8bbea86d87490
func InsertSortedInt64(data []int64, el int64) []int64 {
	index := sort.Search(len(data), func(i int) bool { return data[i] > el })
	data = append(data, 0)
	copy(data[index+1:], data[index:])
	data[index] = el
	return data
}
