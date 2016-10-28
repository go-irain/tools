package amqplib

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

func (c *Channel) Consumer() (<-chan amqp.Delivery, error) {
	queue, err := c.channel.QueueDeclare(
		c.queueName, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("QueueDeclare: %s", err)
	}
	log.Printf("declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
		queue.Name, queue.Messages, queue.Consumers, c.key)
	if err := c.channel.QueueBind(
		c.queueName, // name of the queue
		c.key,       // bindingKey
		c.exchange,  // sourceExchange
		false,       // noWait
		nil,         // arguments
	); err != nil {
		c.Close()
		return nil, fmt.Errorf("Queue Bind: %s", err)
	}
	log.Printf("Queue bound to Exchange, starting Consume (consumer tag %q)", c.tag)
	return c.channel.Consume(
		c.queueName, // name
		c.tag,       // consumerTag,
		false,       // noAck
		false,       // exclusive
		false,       // noLocal
		false,       // noWait
		nil,         // arguments
	)
}
