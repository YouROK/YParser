package tasker

import (
	"YParser/utils"
	"time"
)

type Tasker struct {
	Tasks    []*TaskParser
	Threaded bool
	ReadPage func(link string)
}

type TaskParser struct {
	UpdateTime time.Time
	Link       string
	Category   string
	Worker     func(parser *TaskParser)
}

func (t Tasker) Run() {
	if t.Threaded {
		utils.ParallelLimFor(0, len(t.Tasks), 3, func(i int) {
			t.Tasks[i].Worker(t.Tasks[i])
		})
	} else {
		for i := range t.Tasks {
			t.Tasks[i].Worker(t.Tasks[i])
		}
	}
}
