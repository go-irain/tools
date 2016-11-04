package amqplib

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type Msg amqp.Delivery

func (c *Channel) Consumer(queueName, consumerTag string) (<-chan amqp.Delivery, error) {
	queue, err := c.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("QueueDeclare: %s", err)
	}
	log.Printf("declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
		queue.Name, queue.Messages, queue.Consumers, c.key)
	if err := c.channel.QueueBind(
		queueName,  // name of the queue
		c.key,      // bindingKey
		c.exchange, // sourceExchange
		false,      // noWait
		nil,        // arguments
	); err != nil {
		c.Close()
		return nil, fmt.Errorf("Queue Bind: %s", err)
	}
	log.Printf("Queue bound to Exchange, starting Consume (consumer tag %q)", consumerTag)
	return c.channel.Consume(
		queueName,   // name
		consumerTag, // consumerTag,
		false,       // noAck
		false,       // exclusive
		false,       // noLocal
		false,       // noWait
		nil,         // arguments
	)
}
