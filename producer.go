package amqp

import (
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

type Producer struct {
	channel *amqp.Channel
	config  *Config
}

func (p Producer) Publish(msg PublishMessage) (PublishedMessage, error) {
	route, err := p.config.GetRoute(msg.message)
	if err != nil {
		return PublishedMessage{}, err
	}

	prod, err := p.config.GetProducer(route.Producer)
	if err != nil {
		return PublishedMessage{}, err
	}

	var publishing amqp.Publishing
	if msg.msgType == MessageWithReply {
		queue, err := declareQueue(p.channel, "", QueueConfig{Exclusive: true, AutoDelete: true})
		if err != nil {
			return PublishedMessage{}, errors.New("reply to is not defined")
		}

		publishing.ReplyTo = queue.Name
	}

	published, err := publish(p.channel, prod, msg.message, publishing)

	if msg.msgType == MessageWithReply {
		err = p.awaitForReply(publishing.ReplyTo, published, msg.handler)
	}

	return published, err
}

func (p Producer) awaitForReply(replyTo string, published PublishedMessage, handler BodyHandler) error {
	consumerName := fmt.Sprintf("%s%d", replyTo, published.CorrelationID.ID())
	messages, err := consume(p.channel, consumerName, ConsumerConfig{Queue: replyTo})
	if err != nil {
		log.Println(err.Error())

		return err
	}

	go func() {
		for msg := range messages {
			if msg.CorrelationId != published.CorrelationID.String() {
				continue
			}

			handleMessage(p.channel, msg, handler)

			p.channel.Cancel(consumerName, false)

			return
		}
	}()

	return nil
}

func WaitForReply[T interface{}](i T) (*AwaitForBodyHandler, *T) {
	var await AwaitForBodyHandler
	await.add(1)
	await.Handler = func(body Body) Handled {
		body.To(&i)
		await.Done()

		return HandledSuccessfully()
	}

	return &await, &i
}
