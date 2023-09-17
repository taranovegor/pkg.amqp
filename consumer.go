package amqp

import (
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

type acknowledge string

const (
	ack    acknowledge = "ack"
	nack   acknowledge = "nack"
	reject acknowledge = "reject"
)

type Body map[string]interface{}

func (b Body) To(i interface{}) {
	bytes, _ := json.Marshal(b)
	_ = json.Unmarshal(bytes, &i)
}

type Handled struct {
	ack     acknowledge
	requeue bool
	reply   interface{}
}

func (h Handled) WithReply(r interface{}) Handled {
	h.reply = r

	return h
}

func HandledSuccessfully() Handled {
	return Handled{ack, false, nil}
}

func HandledNotSuccessfully(requeue bool) Handled {
	return Handled{nack, requeue, nil}
}

func HandledAndRejected() Handled {
	return Handled{reject, false, false}
}

type Consumer interface {
	Name() string
	Handle(Body) Handled
}

type consumer struct {
	channel *amqp.Channel
	config  *Config
}

func (c consumer) Consume() {
	for name, cfg := range c.config.exchanges {
		declareExchange(c.channel, name, cfg)
	}

	for name, cfg := range c.config.queues {
		declareQueue(c.channel, name, cfg)
	}

	for name, handler := range c.config.handlers {
		cfg, err := c.config.GetConsumer(name)
		if err != nil {
			log.Println(err.Error())

			continue
		}

		messages, err := consume(c.channel, name, cfg)
		if err != nil {
			log.Println(err.Error())

			continue
		}

		go func(h Consumer) {
			for msg := range messages {
				handleMessage(c.channel, msg, func(b Body) Handled {
					return h.Handle(b)
				})
			}
		}(handler)
	}
}

func handleMessage(ch *amqp.Channel, msg amqp.Delivery, handler BodyHandler) {
	log.Printf("received a message for handler: %s", msg.Body)

	body := Body{}
	err := json.Unmarshal(msg.Body, &body)
	if err != nil {
		msg.Reject(true)

		return
	}

	handled := handler(body)
	if msg.ReplyTo != "" {
		reply := handled.reply
		if reply == nil {
			reply = NoReply{}
		}

		publish(ch, ProducerConfig{
			Queues: []string{msg.ReplyTo},
		}, reply, amqp.Publishing{
			CorrelationId: msg.CorrelationId,
		})
	}

	switch handled.ack {
	case nack:
		msg.Nack(false, handled.requeue)
		break
	case reject:
		msg.Reject(handled.requeue)
		break
	default:
		msg.Ack(false)
	}
}
