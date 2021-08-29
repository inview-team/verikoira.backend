package rmq

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type Publisher struct {
	address string
	queue   string
	conn    *amqp.Connection
	channel *amqp.Channel
	done    chan error
}

func NewPublisher(addr, queue string) *Publisher {
	return &Publisher{
		address: addr,
		queue:   queue,
		done:    make(chan error),
	}
}

func (p *Publisher) Send(data []byte) error {
	be := backoff.NewExponentialBackOff()
	be.MaxElapsedTime = time.Minute
	be.InitialInterval = 1 * time.Second
	be.Multiplier = 2
	be.MaxInterval = 15 * time.Second

	b := backoff.WithContext(be, context.Background())
	for {
		d := b.NextBackOff()
		if d == backoff.Stop {
			return fmt.Errorf("stop reconnecting")
		}
		<-time.After(d)
		if err := p.connect(); err != nil {
			zap.L().Debug("could not reconnect", zap.Error(err))

			continue
		}
		err := p.channel.Publish(
			"",
			p.queue,
			false,
			false,
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "application/json",
				Body:         data,
			})
		if err != nil {
			zap.L().Error("failed to send data", zap.Error(err))

			continue
		}
		zap.L().Debug("sent", zap.String("data", string(data)))

		return nil
	}
}

func (p *Publisher) connect() error {
	conn, err := amqp.Dial(p.address)
	if err != nil {
		return err
	}
	p.conn = conn

	p.channel, err = p.conn.Channel()
	if err != nil {
		return err
	}

	go func() {
		zap.L().Debug("closing", zap.Error(<-p.conn.NotifyClose(make(chan *amqp.Error))))
		p.done <- errors.New("channel closed")
	}()

	err = p.channel.ExchangeDeclare(
		p.queue,  // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)

	return err
}

func (p *Publisher) Reconnect() error {
	be := backoff.NewExponentialBackOff()
	be.MaxElapsedTime = time.Minute
	be.InitialInterval = 1 * time.Second
	be.Multiplier = 2
	be.MaxInterval = 15 * time.Second

	b := backoff.WithContext(be, context.Background())
	for {
		d := b.NextBackOff()
		if d == backoff.Stop {
			return fmt.Errorf("stop reconnecting")
		}
		<-time.After(d)
		if err := p.connect(); err != nil {
			zap.L().Debug("could not reconnect", zap.Error(err))
			continue
		}
		return nil
	}
}

func (p *Publisher) Close() error {
	return p.conn.Close()
}
