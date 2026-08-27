package rabbitmq

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	wg   sync.WaitGroup
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

// StartConsuming sets up the DLX, DLQ, retry queues, and starts consuming.
func (c *Consumer) StartConsuming(queueName string, handler func([]byte) (retryable bool, err error)) error {
	// 1. Declare DLX and DLQ
	if err := c.ch.ExchangeDeclare("recipe_dlx", "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("failed to declare dlx: %w", err)
	}
	if _, err := c.ch.QueueDeclare("recipe_dlq", true, false, false, false, nil); err != nil {
		return fmt.Errorf("failed to declare dlq: %w", err)
	}
	if err := c.ch.QueueBind("recipe_dlq", "", "recipe_dlx", false, nil); err != nil {
		return fmt.Errorf("failed to bind dlq: %w", err)
	}

	// 2. Declare Retry Exchange and Queue
	if err := c.ch.ExchangeDeclare("recipe_retry_exchange", "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("failed to declare retry exchange: %w", err)
	}

	retryArgs := amqp.Table{
		"x-dead-letter-exchange":    "",          // route back to default exchange
		"x-dead-letter-routing-key": queueName,   // back to main queue
		"x-message-ttl":             int32(5000), // 5 seconds wait
	}
	if _, err := c.ch.QueueDeclare("recipe_retry_queue", true, false, false, false, retryArgs); err != nil {
		return fmt.Errorf("failed to declare retry queue: %w", err)
	}
	if err := c.ch.QueueBind("recipe_retry_queue", "", "recipe_retry_exchange", false, nil); err != nil {
		return fmt.Errorf("failed to bind retry queue: %w", err)
	}

	// 3. Declare Main Queue
	mainArgs := amqp.Table{
		"x-dead-letter-exchange": "recipe_dlx",
	}
	_, err := c.ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		mainArgs,
	)
	if err != nil {
		return fmt.Errorf("failed to declare main queue: %w", err)
	}

	msgs, err := c.ch.Consume(
		queueName,
		"worker_consumer", // consumer tag
		false,             // auto-ack (we do manual ack)
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	go func() {
		for d := range msgs {
			c.wg.Add(1)
			retryable, err := handler(d.Body)
			if err == nil {
				d.Ack(false)
				c.wg.Done()
				continue
			}

			if !retryable {
				log.Printf("Permanent error processing message: %v. Sending to DLX.", err)
				d.Nack(false, false) // goes to DLX
				c.wg.Done()
				continue
			}

			// Handle Transient Errors
			retryCount := int32(0)
			if count, ok := d.Headers["x-retry-count"].(int32); ok {
				retryCount = count
			}

			if retryCount >= 3 {
				log.Printf("Exceeded max retries (3) for message. Error: %v. Sending to DLX.", err)
				d.Nack(false, false) // exhausted, send to DLX
			} else {
				log.Printf("Transient error (attempt %d/3): %v. Requeueing via retry exchange.", retryCount+1, err)

				// Ensure headers map exists
				headers := d.Headers
				if headers == nil {
					headers = make(amqp.Table)
				}
				headers["x-retry-count"] = retryCount + 1

				errPublish := c.ch.Publish(
					"recipe_retry_exchange",
					"",
					false,
					false,
					amqp.Publishing{
						Headers:      headers,
						ContentType:  d.ContentType,
						Body:         d.Body,
						DeliveryMode: amqp.Persistent,
						Timestamp:    time.Now(),
					},
				)

				if errPublish != nil {
					log.Printf("Failed to publish to retry exchange: %v. Nacking to DLX as fallback.", errPublish)
					d.Nack(false, false)
				} else {
					// Successfully published to retry exchange, so we ack the original message
					d.Ack(false)
				}
			}
			c.wg.Done()
		}
	}()
	return nil
}

func (c *Consumer) StopConsuming() error {
	return c.ch.Cancel("worker_consumer", false)
}

func (c *Consumer) Wait() {
	c.wg.Wait()
}
