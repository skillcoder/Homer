package helper

/* vim: set ts=2 sw=2 sts=2 ff=unix ft=go noet: */

import (
	"reflect"
	"testing"
	//"time"
	"fmt"
)

// 2 октября 2018 г. 16:32:59
var timestamp = int64(1538487179)

func BenchmarkGetPrevTimestampLoop(b *testing.B) {
	for n := 0; n < b.N; n++ {
		GetPrevTimestampLoop(timestamp, 60)
	}
}

func BenchmarkGetPrevTimestamp(b *testing.B) {
	for n := 0; n < b.N; n++ {
		GetPrevTimestamp(timestamp, 60)
	}
}

func BenchmarkGetNextMinute(b *testing.B) {
	for n := 0; n < b.N; n++ {
		GetNextMinute(timestamp)
	}
}

// AssertEqual checks if values are equal
func AssertEqual(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		return
	}
	// debug.PrintStack()
	t.Errorf("Received %v (type %v), expected %v (type %v)", a, reflect.TypeOf(a), b, reflect.TypeOf(b))
}

func TestGetPrevTimestamp(t *testing.T) {
	fmt.Println("Test:", timestamp)
	var prev = GetPrevTimestamp(timestamp, 60)
	fmt.Println("Prev:", prev)
	AssertEqual(t, prev, GetPrevTimestampLoop(timestamp, 60))
}
