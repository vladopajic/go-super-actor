package example_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vladopajic/go-actor/actor"

	"github.com/vladopajic/go-super-actor"
	. "github.com/vladopajic/go-super-actor/example"
)

func Test_PizzaBaker(t *testing.T) {
	t.Parallel()

	t.Run("actor", func(t *testing.T) {
		t.Parallel()

		testPizzaBaker(t, NewPizzaBaker)
	})

	t.Run("worker", func(t *testing.T) {
		t.Parallel()

		bakeReqMailbox := actor.NewMailbox[BakeRequest]()
		bakeReqMailbox.Start()
		t.Cleanup(bakeReqMailbox.Stop)

		testPizzaBaker(t, func() PizzaBaker {
			return NewPizzaBakeWorker(bakeReqMailbox)
		})
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
			Topings: []Topping{"ketchup", "bacon", "salami", "origaon", "mushrooms"},
		})
		assert.Equal(t, actor.WorkerContinue, sa.DoWork())
		assert.NoError(t, (<-respC).Error)
	}

	{ // Invalid bake request
		respC := baker.Bake(PizzaBakeRequest{
			Topings: []Topping{"ketchup", "bacon", "salami", "strawberry"},
		})
		assert.Equal(t, actor.WorkerContinue, sa.DoWork())
		assert.Error(t, (<-respC).Error)
	}
}
