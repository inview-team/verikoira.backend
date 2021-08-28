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

type Consumer struct {
	addr    string
	queue   string
	conn    *amqp.Connection
	channel *amqp.Channel
	done    chan error
}

func NewConsumer(addr, queue string) *Consumer {
	return &Consumer{
		addr:  addr,
		queue: queue,
		done:  make(chan error),
	}
}

func (c *Consumer) connect() error {
	conn, err := amqp.Dial(c.addr)
	if err != nil {
		return err
	}
	c.conn = conn

	c.channel, err = c.conn.Channel()
	if err != nil {
		return err
	}

	go func() {
		zap.L().Debug("closing", zap.Error(<-c.conn.NotifyClose(make(chan *amqp.Error))))
		c.done <- errors.New("channel closed")
	}()

	_, err = c.channel.QueueDeclare(
		c.queue,
		true,
		false,
		false,
		false,
		nil,
	)
	return err
}

func (c *Consumer) Reconnect(ctx context.Context) (<-chan amqp.Delivery, error) {
	be := backoff.NewExponentialBackOff()
	be.MaxElapsedTime = 3 * time.Minute
	be.InitialInterval = 1 * time.Second
	be.Multiplier = 2
	be.MaxInterval = 30 * time.Second

	b := backoff.WithContext(be, ctx)
	for {
		d := b.NextBackOff()
		if d == backoff.Stop {
			return nil, fmt.Errorf("stop reconnecting")
		}

		select {
		case <-ctx.Done():
			return nil, nil
		case <-time.After(d):
			if err := c.connect(); err != nil {
				zap.L().Debug("could not connect in reconnect call", zap.Error(err))

				continue
			}
			msgs, err := c.channel.Consume(
				c.queue,
				"",
				false,
				false,
				false,
				false,
				nil,
			)
			if err != nil {
				zap.L().Debug("could not connect", zap.Error(err))

				continue
			}

			return msgs, nil
		}
	}
}

func (c *Consumer) Close() error {
	err := c.channel.Close()
	if err != nil {
		return err
	}

	return c.conn.Close()
}
