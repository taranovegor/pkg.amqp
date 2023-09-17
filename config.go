package amqp

import (
	"errors"
	"reflect"
)

var ConfigNotFound = errors.New("config not found")

type ConsumerConfig struct {
	Queue     string
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Args      map[string]interface{}
}

type ExchangeConfig struct {
	Kind       string
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       map[string]interface{}
}

type QueueConfig struct {
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       map[string]interface{}
}

type ProducerConfig struct {
	Exchange  string
	Key       string
	ReplyTo   string
	Mandatory bool
	Immediate bool
}

type RouteConfig struct {
	Type     interface{}
	Producer string
}

type Config struct {
	appId     string
	consumers map[string]ConsumerConfig
	exchanges map[string]ExchangeConfig
	queues    map[string]QueueConfig
	producers map[string]ProducerConfig
	routing   map[string]RouteConfig
	handlers  map[string]Consumer
}

func NewConfig(
	consumers map[string]ConsumerConfig,
	exchanges map[string]ExchangeConfig,
	queues map[string]QueueConfig,
	producers map[string]ProducerConfig,
	routing map[interface{}]RouteConfig,
) Config {
	rm := map[string]RouteConfig{}
	for i, r := range routing {
		rm[reflect.TypeOf(i).String()] = r
	}

	return Config{
		consumers: consumers,
		exchanges: exchanges,
		queues:    queues,
		producers: producers,
		routing:   rm,
		handlers:  make(map[string]Consumer),
	}
}

func (c Config) GetConsumer(name string) (ConsumerConfig, error) {
	if c, f := c.consumers[name]; f {
		return c, nil
	}

	return ConsumerConfig{}, ConfigNotFound
}

func (c Config) GetExchange(name string) (ExchangeConfig, error) {
	if e, f := c.exchanges[name]; f {
		return e, nil
	}

	return ExchangeConfig{}, ConfigNotFound
}

func (c Config) GetQueue(name string) (QueueConfig, error) {
	if q, f := c.queues[name]; f {
		return q, nil
	}

	return QueueConfig{}, ConfigNotFound
}

func (c Config) GetProducer(name string) (ProducerConfig, error) {
	if p, f := c.producers[name]; f {
		return p, nil
	}

	return ProducerConfig{}, ConfigNotFound
}

func (c Config) GetRoute(i interface{}) (RouteConfig, error) {
	var name string
	switch i.(type) {
	case string:
		name = i.(string)
	default:
		name = reflect.TypeOf(i).String()
	}

	if r, f := c.routing[name]; f {
		return r, nil
	}

	return RouteConfig{}, ConfigNotFound
}
