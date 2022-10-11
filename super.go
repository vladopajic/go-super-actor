package super

import (
	"errors"

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
	workReqC chan chan actor.WorkerStatus
	w        actor.Worker
}

func (s *superWorker) DoWork() actor.WorkerStatus {
	respC := make(chan actor.WorkerStatus)
	s.workReqC <- respC

	return <-respC
}

func (s *superWorker) Start() {
	s.workReqC = make(chan chan actor.WorkerStatus)
	go s.handleWorkRequest()
}

func (s *superWorker) Stop() {
	close(s.workReqC)
}

func (s *superWorker) handleWorkRequest() {
	for {
		respC, ok := <-s.workReqC
		if !ok {
			return
		}

		respC <- s.w.DoWork(actor.ContextStarted())
	}
}
