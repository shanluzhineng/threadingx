package timeline

import (
	"sync"
	"time"

	"github.com/abmpio/threadingx/collection"
)

const (
	slots                    = 300
	defaultTimeWheelInterval = time.Millisecond * 16
	// //用于限制当前正在工作的个数
	// schedulerWorkers = 1
)

// 时间轴服务,以每秒60帧的速率轮循(即16毫秒一次轮循)
type ITimeline interface {
	// 订阅时间轴轮循通知，以用来接收时间轴通知，这里订阅的将会一直工作在这个时间轴中，直到调用<see cref="Unsubscribe(ITimelineObserver)"/>方法取消订阅为止
	Subscribe(timelineObserver ITimelineObserver) ITimelineObserver

	// 订阅时间轴轮轮循通知，只通知一次，通知到达一次后在下次的时间轴中将不会再次通知，此方法会自动执行取消订阅
	SubscribeAsOnTime(timelineObserver ITimelineObserver, delayTime time.Duration) error

	// 取消原有的订阅
	Unsubscribe(timelineObserver ITimelineObserver) error
}

// 主timeline
var (
	_ ITimeline = (*timeline)(nil)

	//默认的timeline
	defaultTimeline ITimeline
	once            sync.Once
)

func GetDefaultTimeline() ITimeline {
	once.Do(func() {
		defaultTimeline = newDefaultTimeline()
	})
	return defaultTimeline
}

type timeline struct {
	taskScheduler ITaskScheduler

	//只执行一次的observer队列
	registedOneTimeObserverQueue *collection.Queue
	//一直订阅的observer列表
	registedObserverList []ITimelineObserver
	rwLock               sync.RWMutex

	isChanged bool
}

func newDefaultTimeline() ITimeline {
	timelineService := &timeline{
		registedOneTimeObserverQueue: collection.NewQueue(5),
		registedObserverList:         make([]ITimelineObserver, 0),
		isChanged:                    false,
	}
	taskScheduler := NewTaskScheduler()
	timelineService.taskScheduler = taskScheduler
	return timelineService
}

// #region ITimeline Members

// 订阅时间轴轮轮循通知，以用来接收时间轴通知，这里订阅的将会一直工作在这个时间轴中，直到调用Unsubscribe(ITimelineObserver)方法取消订阅为止
func (s *timeline) Subscribe(timelineObserver ITimelineObserver) ITimelineObserver {

	s.rwLock.Lock()
	s.registedObserverList = append(s.registedObserverList, timelineObserver)
	s.isChanged = true
	s.rwLock.Unlock()

	return timelineObserver
}

// 订阅时间轴轮轮循通知，只通知一次，通知到达一次后在下次的时间轴中将不会再次通知，此方法会自动执行取消订阅
func (s *timeline) SubscribeAsOnTime(timelineObserver ITimelineObserver, delayTime time.Duration) error {

	if delayTime == 0 {
		//立即执行，不延时
	}
	return nil
}

// 取消原有的订阅
func (s *timeline) Unsubscribe(timelineObserver ITimelineObserver) error {
	return nil
}

// #endregion
