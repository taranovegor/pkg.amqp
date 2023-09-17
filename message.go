package amqp

import (
	"github.com/google/uuid"
	"sync"
	"time"
)

const (
	MessageRegular   MessageType = "regular"
	MessageWithReply MessageType = "with_reply"
)

type MessageType string
type BodyHandler func(Body) Handled
type AwaitForBodyHandler struct {
	wg      sync.WaitGroup
	Handler BodyHandler
}

func (h *AwaitForBodyHandler) add(delta int) {
	h.wg.Add(delta)
}

func (h *AwaitForBodyHandler) Wait() {
	h.wg.Wait()
}

func (h *AwaitForBodyHandler) Done() {
	h.wg.Done()
}

type PublishMessage struct {
	msgType MessageType
	message interface{}
	handler BodyHandler
}

func MessageToPublish(msg interface{}) PublishMessage {
	return PublishMessage{
		msgType: MessageRegular,
		message: msg,
	}
}

func MessageToPublishWithReply(msg interface{}, handler BodyHandler) PublishMessage {
	return PublishMessage{
		msgType: MessageWithReply,
		message: msg,
		handler: handler,
	}
}

type PublishedMessage struct {
	ID            uuid.UUID
	CorrelationID uuid.UUID
	SentAt        time.Time
	Message       interface{}
}

type NoReply struct {
}
