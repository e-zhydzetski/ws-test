package wsserver

import (
	"fmt"
	"time"
)

var ErrScheduleTimeout = fmt.Errorf("schedule error: timed out")

type Pool struct {
	work chan func()
}

func NewPool(size int) *Pool {
	p := &Pool{
		work: make(chan func()),
	}
	for i := 0; i < size; i++ {
		go p.worker()
	}
	return p
}

func (p *Pool) Schedule(task func()) {
	_ = p.schedule(task, nil)
}

func (p *Pool) ScheduleTimeout(timeout time.Duration, task func()) error {
	return p.schedule(task, time.After(timeout))
}

func (p *Pool) schedule(task func(), timeout <-chan time.Time) error {
	select {
	case <-timeout:
		return ErrScheduleTimeout
	case p.work <- task:
		return nil
	}
}

func (p *Pool) worker() {
	for task := range p.work {
		task()
	}
}
