package lang

import "sync/atomic"

type SafeBool struct {
	val int32
}

func (b *SafeBool) Set(value bool) {
	var i int32
	if value {
		i = 1
	} else {
		i = 0
	}
	atomic.StoreInt32(&b.val, i)
}

func (b *SafeBool) Get() bool {
	return atomic.LoadInt32(&b.val) != 0
}
