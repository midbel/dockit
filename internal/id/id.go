package id

import (
	"sync/atomic"
)

var global uint64

func Next() uint64 {
	return atomic.AddUint64(&global, 1)
}
