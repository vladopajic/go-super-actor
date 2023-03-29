# go-super-actor

[![test](https://github.com/vladopajic/go-super-actor/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/vladopajic/go-super-actor/actions/workflows/test.yml)
[![lint](https://github.com/vladopajic/go-super-actor/actions/workflows/lint.yml/badge.svg?branch=main)](https://github.com/vladopajic/go-super-actor/actions/workflows/lint.yml)

`go-super-actor` (or just `super`) is addon abstraction for [go-actor](https://github.com/vladopajic/go-actor) designed for testing actors and workers using same testing logic.

## Example 

See example of `go-super-actor` (runnable code is in [example](./example/) folder).

First we need actor and worker that needs to be tested. In this example we have `PizzaBaker` actor and worker.

``` go
type PizzaBaker interface {
	Bake(req PizzaBakeRequest) <-chan PizzaBakeResponse
}

type PizzaBakerActor interface {
	actor.Actor
	PizzaBaker
}

type PizzaBakeRequest struct {
	Toppings []Topping
}

type PizzaBakeResponse struct {
	Error   error
	BakedAt time.Time
}

func NewPizzaBaker() PizzaBakerActor {
	bakeReqMailbox := actor.NewMailbox[bakeRequest]()
	w := newPizzaBakeWorker(bakeReqMailbox)

	return &pizzaBakerActor{
		Actor:      actor.Combine(actor.New(w), bakeReqMailbox),
		PizzaBaker: w,
	}
}

type pizzaBakerActor struct {
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
	if invalidToping := FilterInvalidToping(wreq.req.Toppings); len(invalidToping) > 0 {
		wreq.respC <- PizzaBakeResponse{
			Error: fmt.Errorf("failed to bake pizza: invalid topping requested %+s", invalidToping),
		}

		return
	}

	wreq.respC <- PizzaBakeResponse{
		BakedAt: time.Now(),
	}
}
```

Now we can test `NewPizzaBaker()` and `newPizzaBakeWorker(...)` using same testing logic `testPizzaBaker(...)`.

```go
func Test_PizzaBaker(t *testing.T) {
	t.Parallel()

	t.Run("actor", func(t *testing.T) {
		t.Parallel()

		testPizzaBaker(t, NewPizzaBaker)
	})

	t.Run("worker", func(t *testing.T) {
		t.Parallel()

		fact := func() PizzaBaker {
			bakeReqMailbox := actor.NewMailbox[BakeRequest]()
			bakeReqMailbox.Start()
			t.Cleanup(bakeReqMailbox.Stop)

			return NewPizzaBakeWorker(bakeReqMailbox)
		}

		testPizzaBaker(t, fact)
	})
}

type factoryFn[T PizzaBaker] func() T

func testPizzaBaker[T PizzaBaker](t *testing.T, fact factoryFn[T]) {
	t.Helper()

	baker := fact()
	sa, err := super.New(baker)
	assert.NoError(t, err)

	sa.Start()
	defer sa.Stop()

	{ //  Valid bake request
		respC := baker.Bake(PizzaBakeRequest{
			Toppings: []Topping{"ketchup", "bacon", "salami", "oregano", "mushrooms"},
		})
		assert.Equal(t, actor.WorkerContinue, sa.DoWork())
		assert.NoError(t, (<-respC).Error)
	}

	{ // Invalid bake request
		respC := baker.Bake(PizzaBakeRequest{
			Toppings: []Topping{"ketchup", "bacon", "salami", "strawberry"},
		})
		assert.Equal(t, actor.WorkerContinue, sa.DoWork())
		assert.Error(t, (<-respC).Error)
	}
}
```