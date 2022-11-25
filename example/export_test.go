package example

import "github.com/vladopajic/go-actor/actor"

type BakeRequest = bakeRequest

func NewPizzaBakeWorker(
	bakeReqMailbox actor.Mailbox[BakeRequest],
) *pizzaBakeWorker {
	return newPizzaBakeWorker(bakeReqMailbox)
}
