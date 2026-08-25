package rabbitmq

import (
	"fmt"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewConsumer() (*Consumer, error) {
	host := os.Getenv("RABBITMQ_HOST")
	port := os.Getenv("RABBITMQ_PORT")
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5673"
	}

	url := fmt.Sprintf("amqp://foodie_mq:foodie_mq_pass@%s:%s/", host, port)

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return &Consumer{conn: conn, ch: ch}, nil
}

func (c *Consumer) Close() {
	if c.ch != nil {
		c.ch.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *Consumer) StartConsuming(queueName string, handler func([]byte) error) error {
	_, err := c.ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %w", err)
	}

	msgs, err := c.ch.Consume(
		queueName,
		"",    // consumer
		false, // auto-ack (we do manual ack)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	go func() {
		for d := range msgs {
			err := handler(d.Body)
			if err != nil {
				log.Printf("Error processing message: %v\n", err)
				d.Nack(false, false)
			} else {
				d.Ack(false)
			}
		}
	}()
	return nil
}
