package threadingx

import (
	"fmt"
	"time"

	"github.com/abmpio/threadingx/collection"
)

const (
	slots                    = 300
	timeWheelInterval        = time.Millisecond * 16
	taskSchedulerKey  string = "threadingx::task_scheduler"
)

type (
	ITaskWorkItem interface {
		DoWork()
	}

	taskWorkItemByFunc struct {
		workItemFunc func()
	}

	TaskScheduler struct {
		shutdown     bool
		workItemList []ITaskWorkItem
		timingWheel  *collection.TimingWheel
	}
)

func (workItem *taskWorkItemByFunc) DoWork() {
	if workItem.workItemFunc == nil {
		return
	}
	workItem.workItemFunc()
}

func newFuncWorkItem(workItemFunc func()) ITaskWorkItem {
	return &taskWorkItemByFunc{
		workItemFunc: workItemFunc,
	}
}

func NewTaskScheduler() (*TaskScheduler, error) {
	scheduler := &TaskScheduler{
		workItemList: make([]ITaskWorkItem, 0),
		shutdown:     false,
	}
	timingWheel, err := collection.NewTimingWheel(timeWheelInterval, slots, scheduler.wakeUpWorkItems)
	if err != nil {
		return nil, err
	}
	scheduler.timingWheel = timingWheel
	scheduler.timingWheel.SetTimer(taskSchedulerKey, nil, timeWheelInterval)
	return scheduler, nil
}

func (s *TaskScheduler) AttatchWorkItem(workItem ITaskWorkItem) error {
	if s.shutdown {
		return fmt.Errorf("taskScheduler is shutdown")
	}
	if workItem == nil {
		return fmt.Errorf("workItem cannot be nil")
	}

	s.workItemList = append(s.workItemList, workItem)
	return nil
}

func (s *TaskScheduler) Start() error {
	s.shutdown = false
	return nil
}

func (s *TaskScheduler) Shutdown() error {
	s.shutdown = true
	return nil
}

func (s *TaskScheduler) AttachFunc(workItemFunc func()) error {
	if s.shutdown {
		return fmt.Errorf("taskScheduler is shutdown")
	}
	if workItemFunc == nil {
		return fmt.Errorf("workItemFunc cannot be nil")
	}
	s.workItemList = append(s.workItemList, newFuncWorkItem(workItemFunc))
	return nil
}

func (s *TaskScheduler) wakeUpWorkItems(k, v interface{}) {
	if s.shutdown {
		return
	}
	for _, eachWorkItem := range s.workItemList {
		if eachWorkItem == nil {
			continue
		}
		eachWorkItem.DoWork()
	}
}
