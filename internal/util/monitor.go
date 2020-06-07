package util

import (
	"context"
	"log"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

func NewMonitor(ctx context.Context, g *errgroup.Group) *Monitor {
	m := &Monitor{
		ctx:    ctx,
		wg:     &sync.WaitGroup{},
		events: make(chan event),
	}

	g.Go(m.collector)
	return m
}

type Monitor struct {
	ctx    context.Context
	wg     *sync.WaitGroup
	events chan event
}

func (m *Monitor) Connect() {
	if m.ctx.Err() != nil {
		return
	}
	m.wg.Add(1)
	defer m.wg.Done()
	m.events <- event{
		Connect: true,
	}
}

func (m *Monitor) Write() {
	if m.ctx.Err() != nil {
		return
	}
	m.wg.Add(1)
	defer m.wg.Done()
	m.events <- event{
		Write: true,
	}
}

func (m *Monitor) Read() {
	if m.ctx.Err() != nil {
		return
	}
	m.wg.Add(1)
	defer m.wg.Done()
	m.events <- event{
		Read: true,
	}
}

func (m *Monitor) Error(err error) {
	if m.ctx.Err() != nil {
		return
	}
	m.wg.Add(1)
	defer m.wg.Done()
	m.events <- event{
		Err: err,
	}
}

type event struct {
	Connect bool
	Write   bool
	Read    bool
	Err     error
}

func (m *Monitor) collector() error {
	var writes, reads, connects int
	errs := map[string]int{}
	var first time.Time
	var interval time.Duration

	logAndRefresh := func() {
		if first.IsZero() {
			return
		}
		first = time.Time{}
		log.Println("Interval:", interval, "; Connects:", connects, "; Writes:", writes, "; Reads:", reads)
		if len(errs) != 0 {
			log.Println("Errors:")
			for s, c := range errs {
				log.Println(c, " x ", s)
			}
		}
		connects = 0
		writes = 0
		reads = 0
		errs = map[string]int{}
	}

	go func() {
		<-m.ctx.Done()
		m.wg.Wait() // wait all events have been written
		close(m.events)
	}()

	for {
		select {
		case e, ok := <-m.events:
			if !ok { // channel closed
				logAndRefresh()
				return m.ctx.Err()
			}
		repeat:
			if first.IsZero() {
				first = time.Now()
			}
			interval = time.Since(first)
			if interval.Seconds() > 10 {
				logAndRefresh()
				goto repeat
			}
			if e.Err != nil {
				str := e.Err.Error()
				cnt := errs[str]
				errs[str] = cnt + 1
				continue
			}
			if e.Connect {
				connects++
				continue
			}
			if e.Write {
				writes++
				continue
			}
			if e.Read {
				reads++
				continue
			}
			log.Println("inconsistent event", e)
		case <-time.After(1 * time.Second):
			logAndRefresh()
		}
	}
}
