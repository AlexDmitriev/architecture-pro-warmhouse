package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"heating/db"
	"heating/models"

	amqp "github.com/rabbitmq/amqp091-go"
)

// SensorMessage сообщение от монолита.
type SensorMessage struct {
	SensorID     string `json:"sensor_id"`
	SerialNumber string `json:"serial_number"`
	Type         string `json:"type"`
}

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	db      *db.DB
	queue   string
}

// NewConsumer создаёт консьюмер очереди новых сенсоров.
func NewConsumer(rabbitmqURL, queue string, database *db.DB) (*Consumer, error) {
	conn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	if _, err := ch.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &Consumer{
		conn:    conn,
		channel: ch,
		db:      database,
		queue:   queue,
	}, nil
}

// Start слушает очередь и создаёт запись sensor_heating.
func (c *Consumer) Start(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		c.queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("channel closed")
			}
			if err := c.handleMessage(ctx, msg); err != nil {
				log.Printf("sensor event failed: %v", err)
				msg.Nack(false, true)
			} else {
				msg.Ack(false)
			}
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, msg amqp.Delivery) error {
	var payload SensorMessage
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}
	if payload.SerialNumber == "" {
		return fmt.Errorf("serial_number is empty")
	}

	_, err := c.db.UpdateHeating(ctx, models.HeatingUpdate{
		SerialNumber: payload.SerialNumber,
		Temperature:  0,
	})
	if err != nil {
		return fmt.Errorf("save heating sensor: %w", err)
	}
	return nil
}

// Close освобождает соединения.
func (c *Consumer) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

