package amqplib

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

const (
	ExType_Fanout = "fanout"
	ExType_Direct = "direct"
)

type Channel struct {
	channel *amqp.Channel

	exchange     string
	exchangeType string
	key          string

	pushdata chan string
}

func NewChannel(amqpuri, exchange, exchangeType, key string, onClose func(string)) (*Channel, error) {
	c := &Channel{
		exchange:     exchange,
		exchangeType: exchangeType,
		key:          key,
		pushdata:     make(chan string),
	}

	log.Printf("dialing %q", amqpuri)
	conn, err := amqp.Dial(amqpuri)
	if err != nil {
		return nil, fmt.Errorf("Dial: %s", err)
	}

	go func() {
		onClose(fmt.Sprintf("%s", <-conn.NotifyClose(make(chan *amqp.Error))))
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

func (c *Channel) Cancel(queueName string) error {
	return c.channel.Cancel(queueName, false)
}
