package rabbitmq_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/NivRave/socialfoodie/backend/internal/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRabbitMQIntegration(t *testing.T) {
	ctx := context.Background()

	// 1. Spin up RabbitMQ
	req := testcontainers.ContainerRequest{
		Image:        "rabbitmq:3-management",
		ExposedPorts: []string{"5672/tcp"},
		Env: map[string]string{
			"RABBITMQ_DEFAULT_USER": "foodie_mq",
			"RABBITMQ_DEFAULT_PASS": "foodie_mq_pass",
		},
		WaitingFor: wait.ForLog("Server startup complete").WithStartupTimeout(30 * time.Second),
	}
	rmqContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer func() {
		if err := rmqContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	host, err := rmqContainer.Host(ctx)
	require.NoError(t, err)
	port, err := rmqContainer.MappedPort(ctx, "5672")
	require.NoError(t, err)

	os.Setenv("RABBITMQ_HOST", host)
	os.Setenv("RABBITMQ_PORT", port.Port())

	// 2. Test Connection
	consumer, err := rabbitmq.NewConsumer()
	require.NoError(t, err)
	defer consumer.Close()

	// 3. Test publish and consume
	ch, err := amqp.Dial("amqp://foodie_mq:foodie_mq_pass@" + host + ":" + port.Port() + "/")
	require.NoError(t, err)
	defer ch.Close()
	pubCh, err := ch.Channel()
	require.NoError(t, err)
	defer pubCh.Close()

	testArgs := amqp.Table{
		"x-dead-letter-exchange": "recipe_dlx",
	}
	_, err = pubCh.QueueDeclare("test_queue", true, false, false, false, testArgs)
	require.NoError(t, err)

	msgProcessed := make(chan string)

	// Consume
	err = consumer.StartConsuming("test_queue", func(body []byte) (bool, error) {
		msgProcessed <- string(body)
		return false, nil
	})
	require.NoError(t, err)

	// Publish
	err = pubCh.Publish(
		"",
		"test_queue",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte("hello world"),
		},
	)
	require.NoError(t, err)

	// Wait for processing
	select {
	case msg := <-msgProcessed:
		assert.Equal(t, "hello world", msg)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}
