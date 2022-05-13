package sheduler

import (
	"time"
)

type Sheduler struct {
	isWork    bool
	isRun     bool
	everyHour int
	nextStart time.Time
	worker    func()
}

func NewSheduler(everyHour int, worker func()) *Sheduler {
	sh := new(Sheduler)
	sh.everyHour = everyHour
	sh.nextStart = time.Now().Add(time.Duration(everyHour) * time.Hour)
	sh.worker = worker
	return sh
}

func (s *Sheduler) Start() {
	s.isRun = true
	go func() {
		s.do()
		for s.isRun {
			//Пауза пока не наступит время, может смещаться на 20 секунд
			for time.Now().Before(s.nextStart) {
				time.Sleep(time.Second * 20)
				if !s.isRun {
					return
				}
			}
			//Запуск воркера
			s.do()
		}
	}()
}

func (s *Sheduler) do() {
	s.nextStart = time.Now().Add(time.Duration(s.everyHour) * time.Hour)
	if s.isWork {
		return
	}
	s.isWork = true
	s.worker()
	s.isWork = false
}

func (s *Sheduler) Stop() {
	s.isRun = false
}
