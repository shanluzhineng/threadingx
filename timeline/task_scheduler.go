package timeline

import (
	"fmt"
	"sync"
	"time"

	"github.com/abmpio/threadingx/collection"
	"github.com/abmpio/threadingx/timingwheel"

	"github.com/abmpio/threadingx/stringx"
)

type timeIntervalScheduler struct {
	interval time.Duration
}

func (s *timeIntervalScheduler) Next(prev time.Time) time.Time {
	return prev.Add(s.interval)
}

type TaskItem struct {
	key   string
	Value interface{}
}

func NewTaskItem() *TaskItem {
	return &TaskItem{}
}

// 处理taskScheduler的结果类
type TaskItemCompleted struct {
	TaskItem *TaskItem
	//如果panic了，这里是panic的值
	PanicError interface{}
}

type ITaskScheduler interface {
	//多久后执行回调
	AfterFunc(d time.Duration, taskItem TaskItem, callback func(*TaskItem), completeHooks ...func(*TaskItemCompleted)) ITaskSchedulerObserver

	// 调度一个函数，此函数按照interval时间定期执行
	//返回用于此任务的调度key
	SchedulerFunc(interval time.Duration, taskItem TaskItem, callback func(*TaskItem), completeOpts ...func(*TaskItemCompleted)) ITaskSchedulerObserver

	//停止指定的调度项,如果key不存在，则返回false
	StopScheduler(key string) bool
}

type ITaskSchedulerObserver interface {
	GetKey() string
	Stop() bool
}

type taskSchedulerObserver struct {
	timer *timingwheel.Timer
	host  *taskScheduler
}

var _ ITaskSchedulerObserver = (*taskSchedulerObserver)(nil)

func newTaskSchedulerObserver(host *taskScheduler, timer *timingwheel.Timer) ITaskSchedulerObserver {
	return &taskSchedulerObserver{
		host:  host,
		timer: timer,
	}
}

// #region ITaskSchedulerObserver Members

func (o *taskSchedulerObserver) GetKey() string {
	return o.timer.GetKey()
}

func (o *taskSchedulerObserver) Stop() bool {
	result := o.timer.Stop()
	if result {
		o.host.schedulerObserverList.Del(o.GetKey())
	}
	return result
}

// #endregion

type taskScheduler struct {
	timingWheel *timingwheel.TimingWheel

	rwLock                sync.RWMutex
	schedulerObserverList *collection.SafeMap
}

var _ ITaskScheduler = (*taskScheduler)(nil)

func NewTaskScheduler() ITaskScheduler {
	scheduler := &taskScheduler{}
	scheduler.timingWheel = timingwheel.NewTimingWheel(time.Millisecond, slots)
	scheduler.timingWheel.Start()
	return scheduler
}

// #region ITaskScheduler Members

// 调度一个函数
func (s *taskScheduler) AfterFunc(d time.Duration, taskItem TaskItem, callback func(*TaskItem), completeOpts ...func(*TaskItemCompleted)) ITaskSchedulerObserver {

	taskItem.key = s.newKey()
	taskItemCompleted := &TaskItemCompleted{
		TaskItem: &taskItem,
	}
	t := s.timingWheel.AfterFunc(d, func() {
		//执行完成后删除key
		s.schedulerObserverList.Del(taskItem.key)
		defer func() {
			if err := recover(); err != nil {
				fmt.Print(err)
				//保存值
				taskItemCompleted.PanicError = err
			}
			for _, eachHook := range completeOpts {
				eachHook(taskItemCompleted)
			}
		}()
		//触发回调
		if callback != nil {
			callback(&taskItem)
		}
	}).SetKey(taskItem.key)
	observer := newTaskSchedulerObserver(s, t)
	//增加到待执行的列表中
	s.schedulerObserverList.Set(t.GetKey(), observer)
	return observer
}

// 调度一个函数，此函数按照interval时间定期执行
// 返回用于此任务的调度key
func (s *taskScheduler) SchedulerFunc(interval time.Duration,
	taskItem TaskItem,
	callback func(*TaskItem),
	completeOpts ...func(*TaskItemCompleted)) ITaskSchedulerObserver {

	if len(taskItem.key) <= 0 {
		taskItem.key = s.newKey()
	}
	scheduler := &timeIntervalScheduler{
		interval: interval,
	}
	t := s.timingWheel.ScheduleFuncWith(scheduler, taskItem.key, func() {
		taskItemCompleted := &TaskItemCompleted{
			TaskItem: &taskItem,
		}
		defer func() {
			if err := recover(); err != nil {
				fmt.Print(err)
				//保存值
				taskItemCompleted.PanicError = err
			}
			for _, eachHook := range completeOpts {
				eachHook(taskItemCompleted)
			}
		}()
		//触发回调
		if callback != nil {
			callback(&taskItem)
		}
	})
	if t == nil {
		return nil
	}
	observer := newTaskSchedulerObserver(s, t)
	s.schedulerObserverList.Set(taskItem.key, observer)
	return observer
}

// 移除指定的调度项,如果key不存在，则返回false
func (s *taskScheduler) StopScheduler(key string) bool {
	observerValue, ok := s.schedulerObserverList.Get(key)
	if !ok {
		return false
	}
	observer := observerValue.(ITaskSchedulerObserver)
	observer.Stop()
	return true
}

// #endregion

func (s *taskScheduler) newKey() string {
	key := stringx.Randn(10)

	s.rwLock.Lock()
	defer s.rwLock.Unlock()
	for {
		_, ok := s.schedulerObserverList.Get(key)
		if ok {
			return key
		}
	}
}
