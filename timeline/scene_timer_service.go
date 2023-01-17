package timeline

import (
	"time"
)

type ISceneTimerService interface {
	// 启动一个新的计时器(一次性触发的)
	// delayInterval 延时多久后触发回调
	// timerCallback 回调
	// 返回一个timerId
	StartNewOneTimer(delayInterval time.Duration, timerCallback func(*TaskItem) error) string

	//启动一个新的计时器(一次性触发的)
	// delayInterval 延时多久后触发回调
	// timerCallback 回调
	// data 回调函数所需的数据
	StartNewOneTimerWithData(delayInterval time.Duration, timerCallback func(*TaskItem) error, data interface{}) string

	//启动一个新的计时器(一直在运行的)
	StartRecurNewTimer(timerInterval time.Duration, timerCallback func(*TaskItem) error) string

	//移除一个定时器
	RemoveTimer(timerId string)
}

var _ ISceneTimerService = (*sceneTimerService)(nil)

type sceneTimerService struct {
	taskScheduler ITaskScheduler
}

func NewSceneTimerService() ISceneTimerService {
	service := &sceneTimerService{
		taskScheduler: NewTaskScheduler(),
	}
	return service
}

// #region ISceneTimerService Members

// 启动一个新的计时器(一次性触发的)
func (s *sceneTimerService) StartNewOneTimer(delayInterval time.Duration, timerCallback func(*TaskItem) error) string {

	observer := s.taskScheduler.AfterFunc(delayInterval, *NewTaskItem(), timerCallback)
	return observer.GetKey()
}

// 启动一个新的计时器(一次性触发的)
func (s *sceneTimerService) StartNewOneTimerWithData(delayInterval time.Duration, timerCallback func(*TaskItem) error, data interface{}) string {

	taskItem := NewTaskItem()
	taskItem.Value = data
	observer := s.taskScheduler.AfterFunc(delayInterval, *taskItem, timerCallback)
	return observer.GetKey()
}

// 启动一个新的计时器(一直在运行的)
func (s *sceneTimerService) StartRecurNewTimer(timerInterval time.Duration, timerCallback func(*TaskItem) error) string {

	//增加到调度队列中
	observer := s.taskScheduler.SchedulerFunc(timerInterval, *NewTaskItem(), timerCallback)
	if observer == nil {
		return ""
	}
	return observer.GetKey()
}

func (s *sceneTimerService) RemoveTimer(timerId string) {
	s.taskScheduler.StopScheduler(timerId)
}

// #endregion
