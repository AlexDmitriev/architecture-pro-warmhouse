package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher отвечает за публикацию событий о создании сенсоров.
type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queues  map[string]string
}

// SensorMessage сообщение, отправляемое в микросервисы.
type SensorMessage struct {
	SensorID     string `json:"sensor_id"`
	SerialNumber string `json:"serial_number"`
	Type         string `json:"type"`
}

// NewPublisher создает паблишер и объявляет очереди.
func NewPublisher(rabbitURL string, queueByType map[string]string) (*Publisher, error) {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	for _, q := range queueByType {
		if _, err := ch.QueueDeclare(
			q,
			true,  // durable
			false, // auto-delete
			false, // exclusive
			false, // no-wait
			nil,   // args
		); err != nil {
			ch.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to declare queue %s: %w", q, err)
		}
	}

	return &Publisher{
		conn:    conn,
		channel: ch,
		queues:  queueByType,
	}, nil
}

// PublishSensorCreated отправляет событие о новом сенсоре в нужную очередь.
func (p *Publisher) PublishSensorCreated(ctx context.Context, sensorType string, payload SensorMessage) error {
	queueName, ok := p.queues[sensorType]
	if !ok {
		return fmt.Errorf("unknown sensor type: %s", sensorType)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := p.channel.PublishWithContext(
		ctx,
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Close закрывает соединение и канал.
func (p *Publisher) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}

