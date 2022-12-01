package example

import (
	"fmt"
	"time"

	"github.com/vladopajic/go-actor/actor"
)

type PizzaBaker interface {
	Bake(req PizzaBakeRequest) <-chan PizzaBakeResponse
}

type PizzaBakerActor interface {
	actor.Actor
	PizzaBaker
}

type PizzaBakeRequest struct {
	Topings []Topping
}

type PizzaBakeResponse struct {
	Error   error
	BakedAt time.Time
}

func NewPizzaBakeActor() PizzaBakerActor {
	bakeReqMailbox := actor.NewMailbox[bakeRequest]()
	w := newPizzaBakeWorker(bakeReqMailbox)

	return &bizzaBakerActor{
		Actor:      actor.Combine(actor.New(w), bakeReqMailbox),
		PizzaBaker: w,
	}
}

type bizzaBakerActor struct {
	actor.Actor
	PizzaBaker
}

func newPizzaBakeWorker(
	bakeReqMailbox actor.Mailbox[bakeRequest],
) *pizzaBakeWorker {
	return &pizzaBakeWorker{
		bakeReqMailbox: bakeReqMailbox,
	}
}

type workerRequest[Q any, S any] struct {
	req   Q
	respC chan S
}

type bakeRequest = workerRequest[PizzaBakeRequest, PizzaBakeResponse]

type pizzaBakeWorker struct {
	bakeReqMailbox actor.Mailbox[bakeRequest]
}

func (w *pizzaBakeWorker) DoWork(ctx actor.Context) actor.WorkerStatus {
	select {
	case <-ctx.Done():
		return actor.WorkerEnd
	case wreq := <-w.bakeReqMailbox.ReceiveC():
		w.handleBakeRequest(wreq)
	}

	return actor.WorkerContinue
}

func (w *pizzaBakeWorker) Bake(req PizzaBakeRequest) <-chan PizzaBakeResponse {
	respC := make(chan PizzaBakeResponse, 1)
	w.bakeReqMailbox.SendC() <- bakeRequest{req, respC}

	return respC
}

func (w *pizzaBakeWorker) handleBakeRequest(wreq bakeRequest) {
	if invalidToping := FilterInvalidToping(wreq.req.Topings); len(invalidToping) > 0 {
		wreq.respC <- PizzaBakeResponse{
			//nolint:goerr113 // relax
			Error: fmt.Errorf("failed to bake pizza: invalid topping requested %+s", invalidToping),
		}

		return
	}

	wreq.respC <- PizzaBakeResponse{
		BakedAt: time.Now(),
	}
}
