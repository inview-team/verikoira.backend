package rmq

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/streadway/amqp"
)

type Publisher struct {
	address string
	conn    *amqp.Connection
	channel *amqp.Channel
	done    chan error
}

func NewPublisher(addr string) *Publisher {
	return &Publisher{
		address: addr,
		done:    make(chan error),
	}
}

func (p *Publisher) DeclareQueue(queue string) error {
	_, err := p.channel.QueueDeclare(
		queue,
		false,
		false,
		true,
		false,
		nil,
	)

	return err
}

func (p *Publisher) Send(data []byte, queue string) error {
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
		if err := p.Connect(); err != nil {
			log.Printf("could not reconnect: %+v", err)
			continue
		}
		err := p.channel.Publish(
			"",
			queue,
			false,
			false,
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "application/json",
				Body:         data,
			},
		)
		if err != nil {
			fmt.Printf("failed to send data: %+v", err)
			continue
		}
		log.Printf(" [x] Sent %s", data)
		return nil
	}
}

func (p *Publisher) Connect() error {
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
		log.Printf("closing: %s", <-p.conn.NotifyClose(make(chan *amqp.Error)))
		p.done <- errors.New("channel closed")
	}()

	return nil
}

func (p *Publisher) Close() error {
	err := p.channel.Close()
	if err != nil {
		return err
	}

	return p.conn.Close()
}
