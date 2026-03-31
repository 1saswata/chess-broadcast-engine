package telemetry

import (
	"github.com/rabbitmq/amqp091-go"
)

type AMQPCarrier struct {
	Table amqp091.Table
}

func (ac AMQPCarrier) Get(key string) string {
	v, ok := ac.Table[key]
	if !ok || v == nil {
		return ""
	}
	str, ok := v.(string)
	if !ok {
		return ""
	}
	return str
}

func (ac AMQPCarrier) Set(key string, value string) {
	ac.Table[key] = value
}

func (ac AMQPCarrier) Keys() []string {
	r := make([]string, len(ac.Table))
	i := 0
	for k := range ac.Table {
		r[i] = k
		i++
	}
	return r
}
