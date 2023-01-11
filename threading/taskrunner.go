// this file is copied from https://github.com/zeromicro/go-zero/blob/master/core/threading/taskrunner.go
package threading

import (
	"github.com/abmpio/threadingx/lang"
	"github.com/abmpio/threadingx/rescue"
)

// A TaskRunner is used to control the concurrency of goroutines.
type TaskRunner struct {
	limitChan chan lang.PlaceholderType
}

// NewTaskRunner returns a TaskRunner.
func NewTaskRunner(concurrency int) *TaskRunner {
	return &TaskRunner{
		limitChan: make(chan lang.PlaceholderType, concurrency),
	}
}

// Schedule schedules a task to run under concurrency control.
func (rp *TaskRunner) Schedule(task func()) {
	rp.limitChan <- lang.Placeholder

	go func() {
		defer rescue.Recover(func() {
			<-rp.limitChan
		})

		task()
	}()
}
