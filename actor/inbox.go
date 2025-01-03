package actor

import (
	"context"
	"sync/atomic"
)

const (
	defaultThroughput = 300
	messageBatchSize  = 1024 * 4
)

const (
	stopped int32 = iota
	starting
	idle
	running
)

type Scheduler interface {
	Schedule(fn func())
	Throughput() int
}

type goscheduler int

func (goscheduler) Schedule(fn func()) {
	go fn()
}

func (sched goscheduler) Throughput() int {
	return int(sched)
}

func NewScheduler(throughput int) Scheduler {
	return goscheduler(throughput)
}

type Inboxer interface {
	Start(Processer)
	Send(Envelope)
	Stop()
}

type inbox struct {
	ch         chan Envelope
	proc       Processer
	procStatus int32
	size       int
}

func (i *inbox) Start(p Processer) {
	i.proc = p
	go i.processMessages()
}

func (i *inbox) Stop() {
	if atomic.CompareAndSwapInt32(&i.procStatus, running, stopped) {
		close(i.ch)
		i.proc.Shutdown(context.Background(), func() {})
	}
}

func NewInbox(size int) *inbox {
	return &inbox{
		ch:         make(chan Envelope, size),
		procStatus: stopped,
		size:       size,
	}
}

func (i *inbox) Send(msg Envelope) {
	i.ch <- msg
}

func (i *inbox) processMessages() {
	for msg := range i.ch {
		i.proc.Invoke([]Envelope{msg})
	}
}
