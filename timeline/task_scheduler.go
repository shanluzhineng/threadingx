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

type ITaskScheduler interface {
	//多久后执行回调
	AfterFunc(d time.Duration, taskItem TaskItem, callback func(*TaskItem) error, completeOpts ...func(ITaskSchedulerObserver)) ITaskSchedulerObserver

	// 调度一个函数，此函数按照interval时间定期执行
	//返回用于此任务的调度key
	SchedulerFunc(interval time.Duration, taskItem TaskItem, callback func(*TaskItem) error, completeOpts ...func(ITaskSchedulerObserver)) ITaskSchedulerObserver

	//停止指定的调度项,如果key不存在，则返回false
	StopScheduler(key string) bool
}

type ITaskSchedulerObserver interface {
	GetKey() string
	//获取值
	GetTaskItemValue() interface{}
	//如果回调返回了error,用来获取其error
	Error() error
	//如果回调panic了,这里用来获取panic的值
	PanicValue() interface{}

	Stop() bool
}

type taskSchedulerObserver struct {
	timer *timingwheel.Timer
	host  *taskScheduler

	taskItem   *TaskItem
	err        error
	panicValue interface{}
}

var _ ITaskSchedulerObserver = (*taskSchedulerObserver)(nil)

func newTaskSchedulerObserver(host *taskScheduler) *taskSchedulerObserver {

	return &taskSchedulerObserver{
		host: host,
	}
}

// #region ITaskSchedulerObserver Members

func (o *taskSchedulerObserver) GetKey() string {
	return o.timer.GetKey()
}

// 获取值
func (o *taskSchedulerObserver) GetTaskItemValue() interface{} {
	return o.taskItem.Value
}

// 如果回调返回了error,用来获取其error
func (o *taskSchedulerObserver) Error() error {
	return o.err
}

// 如果回调panic了,这里用来获取panic的值
func (o *taskSchedulerObserver) PanicValue() interface{} {
	return o.panicValue
}

func (o *taskSchedulerObserver) Stop() bool {
	if o.timer == nil {
		o.host.schedulerObserverList.Del(o.GetKey())
		return true
	}
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
	scheduler := &taskScheduler{
		schedulerObserverList: collection.NewSafeMap(),
	}
	scheduler.timingWheel = timingwheel.NewTimingWheel(time.Millisecond, slots)
	scheduler.timingWheel.Start()
	return scheduler
}

// #region ITaskScheduler Members

// 调度一个函数
func (s *taskScheduler) AfterFunc(d time.Duration,
	taskItem TaskItem,
	callback func(*TaskItem) error,
	completeOpts ...func(ITaskSchedulerObserver)) ITaskSchedulerObserver {

	taskItem.key = s.newKey()
	observer := newTaskSchedulerObserver(s)
	observer.taskItem = &taskItem

	t := s.timingWheel.AfterFunc(d, func() {
		//执行完成后删除key
		s.schedulerObserverList.Del(taskItem.key)
		defer func() {
			if panicValue := recover(); panicValue != nil {
				fmt.Print(panicValue)
				//保存值
				observer.panicValue = panicValue
			}
			for _, eachHook := range completeOpts {
				eachHook(observer)
			}
		}()
		//触发回调
		if callback != nil {
			err := callback(&taskItem)
			observer.err = err
		}
	}).SetKey(taskItem.key)

	//增加到待执行的列表中
	s.schedulerObserverList.Set(t.GetKey(), observer)
	return observer
}

// 调度一个函数，此函数按照interval时间定期执行
// 返回用于此任务的调度key
func (s *taskScheduler) SchedulerFunc(interval time.Duration,
	taskItem TaskItem,
	callback func(*TaskItem) error,
	completeOpts ...func(ITaskSchedulerObserver)) ITaskSchedulerObserver {

	if len(taskItem.key) <= 0 {
		taskItem.key = s.newKey()
	}
	scheduler := &timeIntervalScheduler{
		interval: interval,
	}
	observer := newTaskSchedulerObserver(s)
	observer.taskItem = &taskItem
	t := s.timingWheel.ScheduleFuncWith(scheduler, taskItem.key, func() {
		defer func() {
			if panicValue := recover(); panicValue != nil {
				fmt.Print(panicValue)
				//保存值
				observer.panicValue = panicValue
			}
		}()
		//触发回调
		if callback != nil {
			err := callback(&taskItem)
			observer.err = err
		}
		for _, eachHook := range completeOpts {
			eachHook(observer)
		}
	})
	if t == nil {
		return nil
	}
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
		if !ok {
			return key
		}
	}
}
