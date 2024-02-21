package timeline

import (
	"sync"
	"time"

	"github.com/abmpio/threadingx/collection"
	"github.com/abmpio/threadingx/threading"
	"github.com/abmpio/threadingx/timingwheel"

	"github.com/abmpio/threadingx/stringx"
)

type timeIntervalScheduler struct {
	interval time.Duration

	stopped bool
}

func (s *timeIntervalScheduler) Next(prev time.Time) time.Time {
	if s.stopped {
		//已经停止
		return time.Time{}
	}
	return prev.Add(s.interval)
}

func (s *timeIntervalScheduler) stop() {
	s.stopped = true
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
	AfterFunc(interval time.Duration, taskItem *TaskItem, callback func(*TaskItem) error, completeOpts ...func(ITaskSchedulerObserver)) ITaskSchedulerObserver

	// 调度一个函数，此函数按照interval时间定期执行,这个定时器的回调是可能会存在着并行执行的
	//返回用于此任务的调度key
	SchedulerFunc(interval time.Duration, taskItem *TaskItem, callback func(*TaskItem) error, completeOpts ...func(ITaskSchedulerObserver)) ITaskSchedulerObserver

	// 调度一个函数，此函数按照interval时间定期执行,这个定时器的回调是不会存在着并行执行的
	// 下一个定时的触发机制是在这个回调执行完成后再开始计时
	//返回用于此任务的调度key
	SchedulerFuncOneByOne(interval time.Duration, taskItem *TaskItem, callback func(*TaskItem) error, completeOpts ...func(ITaskSchedulerObserver)) ITaskSchedulerObserver

	//停止指定的调度项,如果key不存在，则返回false
	StopScheduler(key string) bool
}

type ITaskSchedulerObserver interface {
	GetKey() string
	//获取值
	GetTaskItemValue() interface{}
	//如果回调返回了error,用来获取其error
	Error() error
	AddCompleteCallbacks(callbacks ...func(ITaskSchedulerObserver))
	IsStopped() bool

	Stop() bool
}

type taskSchedulerObserver struct {
	timer     *timingwheel.Timer
	host      *taskScheduler
	scheduler *timeIntervalScheduler

	completeCallbackList []func(ITaskSchedulerObserver)
	taskItem             *TaskItem
	err                  error
}

var _ ITaskSchedulerObserver = (*taskSchedulerObserver)(nil)

func newTaskSchedulerObserver(host *taskScheduler) *taskSchedulerObserver {

	return &taskSchedulerObserver{
		host:                 host,
		completeCallbackList: make([]func(ITaskSchedulerObserver), 0),
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

func (o *taskSchedulerObserver) IsStopped() bool {
	return o.scheduler == nil || o.scheduler.stopped
}

func (o *taskSchedulerObserver) Stop() bool {
	if o.scheduler != nil {
		o.scheduler.stop()
	}

	if o.timer == nil {
		o.host.schedulerObserverList.Del(o.taskItem.key)
		return true
	}
	result := o.timer.Stop()
	if result {
		o.host.schedulerObserverList.Del(o.taskItem.key)
	}
	return result
}

func (o *taskSchedulerObserver) AddCompleteCallbacks(callbacks ...func(ITaskSchedulerObserver)) {
	if callbacks == nil || len(callbacks) <= 0 {
		return
	}
	o.completeCallbackList = append(o.completeCallbackList, callbacks...)
}

func (o *taskSchedulerObserver) notifyCompleted() {
	for _, eachCallback := range o.completeCallbackList {
		eachCallback(o)
	}
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

func (s *taskScheduler) _afterFunc(interval time.Duration, taskItem *TaskItem, callback func(*TaskItem) error, observer *taskSchedulerObserver) {
	if len(taskItem.key) <= 0 {
		taskItem.key = s.newKey()
	}
	t := s.timingWheel.AfterFunc(interval, func() {
		defer func() {
			//执行完成后删除key
			s.schedulerObserverList.Del(taskItem.key)
		}()
		//触发回调
		threading.SafeCallFunc(func() {
			if callback != nil {
				err := callback(taskItem)
				observer.err = err
			}
		})
		threading.SafeCallFunc(observer.notifyCompleted)
	}).SetKey(taskItem.key)
	observer.timer = t
	//增加到待执行的列表中
	s.schedulerObserverList.Set(taskItem.key, observer)
}

// 调度一个函数
func (s *taskScheduler) AfterFunc(interval time.Duration,
	taskItem *TaskItem,
	callback func(*TaskItem) error,
	completeOpts ...func(ITaskSchedulerObserver)) ITaskSchedulerObserver {

	observer := newTaskSchedulerObserver(s)
	observer.taskItem = taskItem
	observer.AddCompleteCallbacks(completeOpts...)

	s._afterFunc(interval, taskItem, callback, observer)
	return observer
}

// 调度一个函数，此函数按照interval时间定期执行
// 返回用于此任务的调度key
func (s *taskScheduler) SchedulerFunc(interval time.Duration,
	taskItem *TaskItem,
	callback func(*TaskItem) error,
	completeOpts ...func(ITaskSchedulerObserver)) ITaskSchedulerObserver {

	if len(taskItem.key) <= 0 {
		taskItem.key = s.newKey()
	}
	scheduler := &timeIntervalScheduler{
		interval: interval,
	}
	observer := newTaskSchedulerObserver(s)
	observer.taskItem = taskItem
	observer.scheduler = scheduler
	observer.AddCompleteCallbacks(completeOpts...)

	t := s.timingWheel.ScheduleFuncWith(scheduler, taskItem.key, func() {
		//触发回调
		threading.SafeCallFunc(func() {
			if callback != nil {
				err := callback(taskItem)
				observer.err = err
			}
		})
		threading.SafeCallFunc(observer.notifyCompleted)
	})
	if t == nil {
		return nil
	}
	s.schedulerObserverList.Set(taskItem.key, observer)
	return observer
}

func (s *taskScheduler) SchedulerFuncOneByOne(interval time.Duration,
	taskItem *TaskItem,
	callback func(*TaskItem) error,
	completeOpts ...func(ITaskSchedulerObserver)) ITaskSchedulerObserver {

	if len(taskItem.key) <= 0 {
		taskItem.key = s.newKey()
	}
	scheduler := &timeIntervalScheduler{
		interval: interval,
	}
	observer := newTaskSchedulerObserver(s)
	observer.taskItem = taskItem
	observer.scheduler = scheduler
	observer.AddCompleteCallbacks(completeOpts...)

	compCallback := func(to ITaskSchedulerObserver) {
		//check IsStopped
		if to.IsStopped() {
			return
		}
		//再次启动
		s._afterFunc(interval, taskItem, callback, observer)
	}
	observer.AddCompleteCallbacks(compCallback)

	s._afterFunc(interval, taskItem, callback, observer)
	return observer
}

// 移除指定的调度项,如果key不存在，则返回false
func (s *taskScheduler) StopScheduler(key string) bool {
	observerValue, ok := s.schedulerObserverList.Get(key)
	if !ok {
		return false
	}
	observer := observerValue.(*taskSchedulerObserver)
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
