package amqp

import (
	"github.com/google/uuid"
	"time"
)

const (
	MessageRegular   MessageType = "regular"
	MessageWithReply MessageType = "with_reply"
)

type MessageType string
type MessageHandlerFunc func(Body) Handled

type PublishMessage struct {
	msgType MessageType
	message interface{}
	handler MessageHandlerFunc
}

func MessageToPublish(msg interface{}) PublishMessage {
	return PublishMessage{
		msgType: MessageRegular,
		message: msg,
	}
}

func MessageToPublishWithReply(msg interface{}, handler MessageHandlerFunc) PublishMessage {
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
