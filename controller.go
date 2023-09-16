package amqp

import amqp "github.com/rabbitmq/amqp091-go"

type Controller struct {
	*consumer
	*producer
}

func Init(
	url string,
	config Config,
	consumers []Consumer,
) (*Controller, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	for _, c := range consumers {
		config.handlers[c.Name()] = c
	}

	return &Controller{
		consumer: &consumer{ch, &config},
		producer: &producer{ch, &config},
	}, nil
}
