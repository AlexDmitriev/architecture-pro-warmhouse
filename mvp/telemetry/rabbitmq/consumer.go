package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"telemetry/db"
	"telemetry/models"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	db      *db.DB
	queue   string
}

// NewConsumer создает новый RabbitMQ consumer
func NewConsumer(rabbitmqURL, queueName string, database *db.DB) (*Consumer, error) {
	conn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Объявляем очередь (если не существует)
	_, err = ch.QueueDeclare(
		queueName, // имя очереди
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &Consumer{
		conn:    conn,
		channel: ch,
		db:      database,
		queue:   queueName,
	}, nil
}

// Message структура сообщения из RabbitMQ
type Message struct {
	SerialNumber string          `json:"serial_number"`
	Value        json.RawMessage `json:"value"`
}

// Start начинает потребление сообщений
func (c *Consumer) Start(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		c.queue, // queue
		"",      // consumer
		false,   // auto-ack (false = ручное подтверждение)
		false,   // exclusive
		false,   // no-local
		false,   // no-wait
		nil,     // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Printf("Waiting for messages from queue: %s", c.queue)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("channel closed")
			}

			if err := c.processMessage(ctx, msg); err != nil {
				log.Printf("Error processing message: %v", err)
				// Отклоняем сообщение (можно настроить retry логику)
				msg.Nack(false, true) // requeue = true для повторной обработки
			} else {
				// Подтверждаем успешную обработку
				msg.Ack(false)
			}
		}
	}
}

// processMessage обрабатывает одно сообщение
func (c *Consumer) processMessage(ctx context.Context, msg amqp.Delivery) error {
	var message Message
	if err := json.Unmarshal(msg.Body, &message); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Валидация
	if message.SerialNumber == "" {
		return fmt.Errorf("serial_number is required")
	}
	if len(message.Value) == 0 {
		return fmt.Errorf("value is required")
	}

	// Создаем запись телеметрии
	telemetry := models.TelemetryCreate{
		SerialNumber: message.SerialNumber,
		Value:        message.Value, // json.RawMessage уже в правильном формате
	}

	// Сохраняем в БД
	_, err := c.db.CreateTelemetry(ctx, telemetry)
	if err != nil {
		return fmt.Errorf("failed to save telemetry to DB: %w", err)
	}

	log.Printf("Telemetry saved: serial_number=%s", message.SerialNumber)
	return nil
}

// Close закрывает соединения
func (c *Consumer) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
