package timeline

type ITimelineObserver interface {
	// 时间轴轮循通知
	// deltaSeconds 与上一次时间轴执行时相关的毫秒值,以ms为单位</param>
	OnNext(deltaSeconds float64)
}

type ActionTimelineObserver struct {
	data interface{}

	nextFunc func()
}

func (o *ActionTimelineObserver) OnNext(deltaSeconds float64) {
	if o.nextFunc == nil {
		return
	}
	o.nextFunc()
}

func NewTimelineObserver(nextFunc func()) ITimelineObserver {
	return &ActionTimelineObserver{
		nextFunc: nextFunc,
	}
}

func NewTimelineObserverWithData(data interface{}, nextFunc func(interface{})) ITimelineObserver {
	newFunc := func() {
		nextFunc(data)
	}
	return &ActionTimelineObserver{
		data:     data,
		nextFunc: newFunc,
	}
}
