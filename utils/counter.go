package utils

import (
	"sync/atomic"
)

type Counter int32

func (c *Counter) Inc() int32 {
	return atomic.AddInt32((*int32)(c), 1)
}

func (c *Counter) Get() int32 {
	return atomic.LoadInt32((*int32)(c))
}
