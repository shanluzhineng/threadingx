package threadingx

type ITimelineObserver interface {
	// 时间轴轮循通知
	// deltaSeconds 与上一次时间轴执行时相关的毫秒值,以ms为单位</param>
	OnNext(deltaSeconds float64)
}
