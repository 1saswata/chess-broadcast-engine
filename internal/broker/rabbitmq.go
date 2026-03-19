package broker

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher struct {
	ch *amqp091.Channel
}

func (rp *RabbitMQPublisher) PublishMove(ctx context.Context, move []byte) error {
	err := rp.ch.PublishWithContext(ctx,
		"chess_broadcast",
		"",
		false,
		false,
		amqp091.Publishing{ContentType: "application/x-protobuf", Body: move},
	)
	return err
}

func NewRabbitMQPublisher(connURL string) (*RabbitMQPublisher, error) {
	rp := &RabbitMQPublisher{}
	conn, err := amqp091.Dial(connURL)
	if err != nil {
		return nil, err
	}
	rp.ch, err = conn.Channel()
	if err != nil {
		return nil, err
	}
	err = rp.ch.ExchangeDeclare(
		"chess_broadcast",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return rp, nil
}
