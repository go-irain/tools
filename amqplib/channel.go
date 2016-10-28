package amqplib

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type Channel struct {
	channel *amqp.Channel

	exchange     string
	exchangeType string
	key          string
	tag          string
	queueName    string
	pushdata     chan string
}

func NewChannel(amqpuri, exchange, exchangeType, key, queuename, ctag string) (*Channel, error) {
	c := &Channel{
		tag:          ctag,
		exchange:     exchange,
		exchangeType: exchangeType,
		key:          key,
		queueName:    queuename,
		pushdata:     make(chan string, 0),
	}

	log.Printf("dialing %q", amqpuri)
	conn, err := amqp.Dial(amqpuri)
	if err != nil {
		return nil, fmt.Errorf("Dial: %s", err)
	}

	go func() {
		fmt.Printf("closing: %s", <-conn.NotifyClose(make(chan *amqp.Error)))
	}()

	log.Printf("got Connection, getting Channel")
	c.channel, err = conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("Channel: %s", err)
	}

	return c, nil
}

func (c *Channel) Close() error {
	return c.channel.Close()
}

func (c *Channel) Cancel() error {
	return c.channel.Cancel(c.queueName, false)
}
