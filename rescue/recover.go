// this file is copied from  https://github.com/zeromicro/go-zero/blob/master/core/rescue/recover.go
package rescue

import (
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

func SafeCallFunc(f func(), cleanups ...func()) {
	defer Recover(cleanups...)

	f()
}
