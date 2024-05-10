package lang

import "sync/atomic"

type SafeInt32 struct {
	val int32
}

func (b *SafeInt32) Set(value int32) {
	atomic.StoreInt32(&b.val, value)
}

func (b *SafeInt32) Get() int32 {
	return atomic.LoadInt32(&b.val)
}
