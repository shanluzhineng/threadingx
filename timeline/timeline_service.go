package timeline

type (
	ITaskWorkItem interface {
		DoWork() error
	}

	taskWorkItemByFunc struct {
		workItemFunc func() error
	}
)

func (workItem *taskWorkItemByFunc) DoWork() error {
	if workItem.workItemFunc == nil {
		return nil
	}
	return workItem.workItemFunc()
}

func newFuncWorkItem(workItemFunc func() error) ITaskWorkItem {
	return &taskWorkItemByFunc{
		workItemFunc: workItemFunc,
	}
}

// func (s *TaskScheduler) AttatchWorkItem(workItem ITaskWorkItem) error {
// 	if workItem == nil {
// 		return fmt.Errorf("workItem cannot be nil")
// 	}

// 	s.workItemList = append(s.workItemList, workItem)
// 	return nil
// }

// func (s *TaskScheduler) AttachFunc(workItemFunc func() error) error {
// 	if workItemFunc == nil {
// 		return fmt.Errorf("workItemFunc cannot be nil")
// 	}
// 	s.workItemList = append(s.workItemList, newFuncWorkItem(workItemFunc))
// 	return nil
// }
