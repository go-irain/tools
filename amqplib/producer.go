package amqplib

import (
	"log"

	"github.com/streadway/amqp"
)

func (c *Channel) Producer() error {
	if err := c.channel.ExchangeDeclare(
		c.exchange,     // name
		c.exchangeType, // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // noWait
		nil,            // arguments
	); err != nil {
		c.Close()
		return err
	}

	for {
		data := <-c.pushdata
		if err := c.channel.Publish(
			c.exchange, // publish to an exchange
			c.key,      // routing to 0 or more queues
			false,      // mandatory
			false,      // immediate
			amqp.Publishing{
				Headers:         amqp.Table{},
				ContentType:     "text/plain",
				ContentEncoding: "",
				Body:            []byte(data),
				DeliveryMode:    amqp.Persistent, // 1=non-persistent, 2=persistent
				Priority:        0,               // 0-9
			},
		); err != nil {
			log.Printf("error:Exchange Publish: %s", err)
			continue
		}
	}
	return nil
}

func (c *Channel) Push(data string) {
	c.pushdata <- data
}
