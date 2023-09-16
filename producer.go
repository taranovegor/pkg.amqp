package amqp

import (
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"sync"
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

	publishing := amqp.Publishing{}
	if msg.msgType == MessageWithReply {
		if prod.ReplyTo == "" {
			return PublishedMessage{}, errors.New("reply to is not defined")
		}

		publishing.ReplyTo = prod.ReplyTo
	}

	published, err := publish(p.channel, prod, msg.message, publishing)
	if msg.msgType == MessageWithReply {
		err = p.awaitForReply(prod.ReplyTo, published, msg.handler)
	}

	return published, err
}

func (p Producer) awaitForReply(replyTo string, published PublishedMessage, handler MessageHandlerFunc) error {
	queue, err := p.config.GetQueue(replyTo)
	if err != nil {
		queue = QueueConfig{Durable: true}
	}
	queue.AutoDelete = false // todo: when sending multiple messages to the queue, a problem may occur

	_, err = declareQueue(p.channel, replyTo, queue)
	if err != nil {
		log.Println(err.Error())

		return err
	}

	consumerName := fmt.Sprintf("%s%d", replyTo, published.CorrelationID.ID())
	messages, err := consume(p.channel, consumerName, ConsumerConfig{Queue: replyTo, NoLocal: false, NoWait: false})
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

func WaitForReply[T interface{}](i T) (MessageHandlerFunc, *sync.WaitGroup, T) {
	var wg *sync.WaitGroup

	return func(body Body) Handled {
		body.To(i)
		wg.Done()

		return HandledSuccessfully()
	}, wg, i
}
