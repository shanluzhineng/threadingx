// this file is copied from  https://github.com/zeromicro/go-zero/blob/master/core/rescue/recover.go
package rescue

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
)

// Recover is used with defer to do cleanup on panics.
// Use it like:
//
//	defer Recover(func() {})
func Recover(cleanups ...func()) {
	for _, cleanup := range cleanups {
		cleanup()
	}

	if p := recover(); p != nil {
		msg := fmt.Sprint(p)
		log.Printf("%s\n%s", msg, debug.Stack())
	}
}

// RecoverCtx is used with defer to do cleanup on panics.
func RecoverCtx(ctx context.Context, cleanups ...func()) {
	for _, cleanup := range cleanups {
		cleanup()
	}

	if p := recover(); p != nil {
		log.Printf("%+v\n%s", p, debug.Stack())
	}
}
