package rabbitmq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"github.com/verizhang/file-manager/internal/config"
	"github.com/verizhang/file-manager/internal/logger"
	"github.com/verizhang/file-manager/internal/messaging"
)

type Client struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewMessaging(cfg config.RabbitMQConfig) (*Client, error) {
	amqpURL := fmt.Sprintf(
		"amqp://%s:%d/",
		cfg.Host,
		cfg.Port,
	)

	conn, err := amqp.DialConfig(
		amqpURL,
		amqp.Config{
			SASL: []amqp.Authentication{
				&amqp.PlainAuth{
					Username: cfg.User,
					Password: cfg.Password,
				},
			},
			Vhost: cfg.VHost,
		},
	)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	return &Client{
		conn: conn,
		ch:   ch,
	}, nil
}

func (c *Client) Publish(
	ctx context.Context,
	topic string,
	payload []byte,
) error {

	_, err := c.ch.QueueDeclare(
		topic,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	return c.ch.PublishWithContext(
		ctx,
		"",
		topic,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        payload,
		},
	)
}

func (c *Client) Subscribe(
	ctx context.Context,
	topic string,
	handler messaging.Handler,
) error {

	_, err := c.ch.QueueDeclare(
		topic,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	msgs, err := c.ch.Consume(
		topic,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case msg, ok := <-msgs:
				if !ok {
					return
				}

				if err := handler(ctx, msg.Body); err != nil {
					logger.Log.Error(
						"failed to process message",
						zap.String("topic", topic),
						zap.Error(err),
					)
					_ = msg.Ack(false)
					continue
				}

				_ = msg.Ack(false)
			}
		}
	}()

	return nil
}

func (c *Client) Close() error {
	if c.ch != nil {
		_ = c.ch.Close()
	}

	if c.conn != nil {
		_ = c.conn.Close()
	}

	return nil
}
