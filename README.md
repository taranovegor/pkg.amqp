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

type LogRegularConsumer struct {
	amqp.Consumer
}

func (h LogRegularConsumer) Name() string {
	return "consumer_log_regular"
}

func (h LogRegularConsumer) Handle(body amqp.Body) amqp.Handled {
	log.Printf("Logged message: %+v", body)

	return amqp.HandledSuccessfully()
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
		"consumer_log_regular": {Queue: "msg.log", Exclusive: false, NoLocal: false, NoWait: false},
		"consumer_regular":     {Queue: "msg.process", Exclusive: false, NoLocal: false, NoWait: false},
		"consumer_with_reply":  {Queue: "msg.with_reply", Exclusive: true, NoLocal: true, NoWait: true},
	},
	map[string]amqp.ExchangeConfig{
		"msg.topic": {Kind: amqp.ExchangeTopic},
	},
	map[string]amqp.QueueConfig{
		"msg.log":        {},
		"msg.process":    {},
		"msg.with_reply": {},
	},
	map[string][]amqp.QueueBindConfig{
		"msg.log":        {{Key: "msg.*", Exchange: "msg.topic"}},
		"msg.process":    {{Key: "msg.regular", Exchange: "msg.topic"}},
		"msg.with_reply": {{Key: "msg.with_reply", Exchange: "msg.topic"}},
	},
	map[string]amqp.ProducerConfig{
		"producer_regular":            {Exchange: "msg.topic", Key: "msg.regular"},
		"producer_awaiting_for_reply": {Exchange: "msg.topic", Key: "msg.with_reply", ReplyTo: "response"},
	},
	map[interface{}]amqp.RouteConfig{
		RegularMessage{}:  {Producer: "producer_regular"},
		MessageForReply{}: {Producer: "producer_awaiting_for_reply"},
	},
)

func main() {
	ctrl, err := amqp.Init("pkg.amqp", os.Getenv("AMQP_URL"), cfg, []amqp.Consumer{
		LogRegularConsumer{},
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
