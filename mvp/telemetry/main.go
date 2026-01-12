package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"telemetry/db"
	"telemetry/rabbitmq"

	"github.com/gin-gonic/gin"
)

func main() {
	// Получаем переменные окружения
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		log.Fatal("RABBITMQ_URL environment variable is required")
	}

	queueName := os.Getenv("RABBITMQ_QUEUE")
	if queueName == "" {
		queueName = "telemetry" // значение по умолчанию
	}

	// Подключаемся к БД
	database, err := db.New(databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Создаем RabbitMQ consumer
	consumer, err := rabbitmq.NewConsumer(rabbitmqURL, queueName, database)
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ consumer: %v", err)
	}
	defer consumer.Close()

	// Создаем контекст с отменой
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем consumer в горутине
	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Printf("Consumer error: %v", err)
		}
	}()

	// Настраиваем HTTP сервер
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// REST API для получения телеметрии
	r.GET("/telemetry", func(c *gin.Context) {
		// Опциональный фильтр по serial_number
		serialNumber := c.Query("serial_number")

		var serialPtr *string
		if serialNumber != "" {
			serialPtr = &serialNumber
		}

		telemetry, err := database.GetTelemetry(c.Request.Context(), serialPtr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get telemetry",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": telemetry,
			"count": len(telemetry),
		})
	})

	// Обработка сигналов для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		cancel()
	}()

	// Запускаем HTTP сервер
	if err := r.Run(":8081"); err != nil {
		log.Printf("HTTP server error: %v", err)
	}
}