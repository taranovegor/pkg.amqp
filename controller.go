package amqp

import amqp "github.com/rabbitmq/amqp091-go"

var appId = "pkg.amqp"

type Controller struct {
	*consumer
	*Producer
}

func Init(
	appName string,
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

	appId = appName
	for _, c := range consumers {
		config.handlers[c.Name()] = c
	}

	return &Controller{
		consumer: &consumer{ch, &config},
		Producer: &Producer{ch, &config},
	}, nil
}
