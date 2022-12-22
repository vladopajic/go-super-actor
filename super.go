package super

import (
	"errors"
	"sync"

	"github.com/vladopajic/go-actor/actor"
)

var ErrNotExpectedObject = errors.New("object must be Actor or Worker")

type Super interface {
	actor.Actor
	DoWork() actor.WorkerStatus
}

func New(aw any) (Super, error) {
	if a, ok := aw.(actor.Actor); ok {
		return &superActor{Actor: a}, nil
	} else if w, ok := aw.(actor.Worker); ok {
		return &superWorker{w: w}, nil
	}

	return nil, ErrNotExpectedObject
}

type superActor struct {
	actor.Actor
}

func (s *superActor) DoWork() actor.WorkerStatus {
	return actor.WorkerContinue
}

type superWorker struct {
	workReqCLock sync.Mutex
	workReqC     chan chan actor.WorkerStatus
	w            actor.Worker
}

func (s *superWorker) DoWork() actor.WorkerStatus {
	respC := make(chan actor.WorkerStatus)

	s.workReqCLock.Lock()
	s.workReqC <- respC
	s.workReqCLock.Unlock()

	return <-respC
}

func (s *superWorker) Start() {
	s.workReqCLock.Lock()
	defer s.workReqCLock.Unlock()

	if s.workReqC != nil {
		return
	}

	s.workReqC = make(chan chan actor.WorkerStatus)
	go handleWorkRequest(s.w, s.workReqC)
}

func (s *superWorker) Stop() {
	s.workReqCLock.Lock()
	defer s.workReqCLock.Unlock()

	if s.workReqC != nil {
		close(s.workReqC)
		s.workReqC = nil
	}
}

func handleWorkRequest(
	w actor.Worker,
	workReqC chan chan actor.WorkerStatus,
) {
	ctx := actor.ContextStarted()

	for {
		respC, ok := <-workReqC
		if !ok {
			return
		}

		respC <- w.DoWork(ctx)
	}
}
