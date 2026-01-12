package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
}

// NewPublisher создает новый RabbitMQ publisher
func NewPublisher(rabbitmqURL, queueName string) (*Publisher, error) {
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

	return &Publisher{
		conn:    conn,
		channel: ch,
		queue:   queueName,
	}, nil
}

// TelemetryMessage структура сообщения для telemetry сервиса
type TelemetryMessage struct {
	SerialNumber string          `json:"serial_number"`
	Value        json.RawMessage `json:"value"`
}

// PublishTelemetry отправляет сообщение в очередь telemetry
func (p *Publisher) PublishTelemetry(serialNumber string, value json.RawMessage) error {
	message := TelemetryMessage{
		SerialNumber: serialNumber,
		Value:        value,
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = p.channel.Publish(
		"",     // exchange
		p.queue, // routing key (имя очереди)
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Message published to queue %s: serial_number=%s", p.queue, serialNumber)
	return nil
}

// Close закрывает соединения
func (p *Publisher) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}
