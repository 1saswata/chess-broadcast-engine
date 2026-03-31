package broker

import (
	"context"

	"github.com/1saswata/chess-broadcast-engine/internal/telemetry"
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
)

type RabbitMQPublisher struct {
	ch   *amqp091.Channel
	conn *amqp091.Connection
}

func (rp *RabbitMQPublisher) PublishMove(ctx context.Context, move []byte) error {
	ctx, span := otel.Tracer("rabbitmq-broker").Start(ctx, "PublishMove")
	defer span.End()
	t := make(amqp091.Table)
	carrier := telemetry.AMQPCarrier{Table: t}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	err := rp.ch.PublishWithContext(ctx,
		"chess_broadcast",
		"",
		false,
		false,
		amqp091.Publishing{ContentType: "application/x-protobuf", Body: move,
			Headers: carrier.Table},
	)
	return err
}

func NewRabbitMQPublisher(connURL string) (*RabbitMQPublisher, error) {
	rp := &RabbitMQPublisher{}
	var err error
	rp.conn, err = amqp091.Dial(connURL)
	if err != nil {
		return nil, err
	}
	rp.ch, err = rp.conn.Channel()
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

func (rp *RabbitMQPublisher) Close() error {
	err := rp.ch.Close()
	if err != nil {
		return err
	}
	err = rp.conn.Close()
	if err != nil {
		return err
	}
	return nil
}
