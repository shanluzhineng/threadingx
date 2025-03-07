// this file is copied from https://github.com/zeromicro/go-zero/blob/master/core/threading/routines.go
package threading

import (
	"bytes"
	"context"
	"runtime"
	"strconv"

	"github.com/shanluzhineng/threadingx/rescue"
)

// GoSafe runs the given fn using another goroutine, recovers if fn panics.
func GoSafe(fn func()) {
	go RunSafe(fn)
}

// GoSafe runs the given fn using another goroutine, recovers if fn panics.
func GoSafeWithCleanup(fn func(), cleanupsFn ...func()) {
	defer rescue.Recover(cleanupsFn...)

	fn()
}

// GoSafeCtx runs the given fn using another goroutine, recovers if fn panics with ctx.
func GoSafeCtx(ctx context.Context, fn func()) {
	go RunSafeCtx(ctx, fn)
}

// RoutineId is only for debug, never use it in production.
func RoutineId() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	// if error, just return 0
	n, _ := strconv.ParseUint(string(b), 10, 64)

	return n
}

// RunSafe runs the given fn, recovers if fn panics.
func RunSafe(fn func(), cleanups ...func()) {
	defer rescue.Recover(cleanups...)

	fn()
}

// RunSafe runs the given fn, recovers if fn panics.
func RunSafeAs[v any](fn func() v, cleanups ...func()) v {
	defer rescue.Recover(cleanups...)

	return fn()
}

// RunSafeCtx runs the given fn, recovers if fn panics with ctx.
func RunSafeCtx(ctx context.Context, fn func()) {
	defer rescue.RecoverCtx(ctx)

	fn()
}

func SafeCallFunc(f func(), cleanups ...func()) {
	defer rescue.Recover(cleanups...)

	f()
}
