# pkg.amqp
Queuing interface package for com projects

## Usage example
```go
package main

import (
	"fmt"
	amqp "github.com/taranovegor/pkg.amqp"
	"log"
	"os"
	"time"
)

type RegularMessage struct {
	Text string
}

type MessageForReply struct {
	Text string
}

type MessageReply struct {
	OriginText string
}

type RegularConsumer struct {
	amqp.Consumer
}

func (h RegularConsumer) Name() string {
	return "consumer_regular"
}

func (h RegularConsumer) Handle(body amqp.Body) amqp.Handled {
	msg := RegularMessage{}
	body.To(&msg)

	log.Printf("Consumed message: %s", msg.Text)

	return amqp.HandledSuccessfully()
}

type WithReplyConsumer struct {
	amqp.Consumer
}

func (h WithReplyConsumer) Name() string {
	return "consumer_with_reply"
}

func (h WithReplyConsumer) Handle(body amqp.Body) amqp.Handled {
	msg := MessageForReply{}
	body.To(&msg)

	log.Printf("Consumed message with reply: %s", msg.Text)

	return amqp.HandledSuccessfully().WithReply(MessageReply{OriginText: msg.Text})
}

var cfg = amqp.NewConfig(
	map[string]amqp.ConsumerConfig{
		"consumer_regular":    {Queue: "regular", Exclusive: false, NoLocal: false, NoWait: false},
		"consumer_with_reply": {Queue: "request", Exclusive: true, NoLocal: true, NoWait: true},
	},
	map[string]amqp.ExchangeConfig{},
	map[string]amqp.QueueConfig{
		"regular": {Durable: false, AutoDelete: false, Exclusive: false, NoWait: false},
		"request": {Durable: true, AutoDelete: true, Exclusive: true, NoWait: true},
	},
	map[string]amqp.ProducerConfig{
		"producer_regular":            {Queues: []string{"regular"}},
		"producer_awaiting_for_reply": {Queues: []string{"request"}, ReplyTo: "response"},
	},
	map[interface{}]amqp.RouteConfig{
		RegularMessage{}:  {Producer: "producer_regular"},
		MessageForReply{}: {Producer: "producer_awaiting_for_reply"},
	},
)

func main() {
	ctrl, err := amqp.Init("pkg.amqp", os.Getenv("AMQP_URL"), cfg, []amqp.Consumer{
		RegularConsumer{},
		WithReplyConsumer{},
	})
	if err != nil {
		panic(err)
	}

	ctrl.Consume()

	ctrl.Publish(
		amqp.MessageToPublish(
			RegularMessage{Text: fmt.Sprintf("regular message, created at %s", time.Now().String())},
		),
	)

	ctrl.Publish(
		amqp.MessageToPublishWithReply(
			MessageForReply{Text: fmt.Sprintf("message for reply, created at %s", time.Now().String())},
			func(body amqp.Body) amqp.Handled {
				msg := MessageReply{}
				body.To(&msg)

				log.Printf("Consumed message reply: %s", msg.OriginText)

				return amqp.HandledSuccessfully()
			},
		),
	)

	select {}
}
```
