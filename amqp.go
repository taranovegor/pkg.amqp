package amqp

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"reflect"
	"time"
)

func declareQueue(ch *amqp.Channel, name string, cfg QueueConfig) (amqp.Queue, error) {
	return ch.QueueDeclare(name, cfg.Durable, cfg.AutoDelete, cfg.Exclusive, cfg.NoWait, cfg.Args)
}

func declareExchange(ch *amqp.Channel, name string, cfg ExchangeConfig) error {
	return ch.ExchangeDeclare(name, cfg.Kind, cfg.Durable, cfg.AutoDelete, cfg.Internal, cfg.NoWait, cfg.Args)
}

func publish(ch *amqp.Channel, cfg ProducerConfig, msg interface{}, pub amqp.Publishing) (PublishedMessage, error) {
	body, err := json.Marshal(msg)
	if err != nil {
		return PublishedMessage{}, err
	}

	sentMsg := PublishedMessage{
		ID:      uuid.New(),
		SentAt:  time.Now(),
		Message: msg,
	}

	if pub.CorrelationId == "" {
		sentMsg.CorrelationID = uuid.New()
	} else {
		sentMsg.CorrelationID, _ = uuid.Parse(pub.CorrelationId)
	}

	pub.ContentType = "application/json"
	pub.CorrelationId = sentMsg.CorrelationID.String()
	pub.MessageId = sentMsg.ID.String()
	pub.Timestamp = sentMsg.SentAt
	pub.Type = reflect.TypeOf(msg).String()
	pub.AppId = appId
	pub.Body = body

	return sentMsg, ch.PublishWithContext(context.Background(), cfg.Exchange, cfg.Key, cfg.Mandatory, cfg.Immediate, pub)
}

func consume(ch *amqp.Channel, name string, cfg ConsumerConfig) (<-chan amqp.Delivery, error) {
	return ch.Consume(cfg.Queue, name, false, cfg.Exclusive, cfg.NoLocal, cfg.NoWait, cfg.Args)
}
