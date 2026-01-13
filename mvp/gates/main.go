package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"gates/db"
	"gates/models"
	"gates/rabbitmq"

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

	sensorQueueName := os.Getenv("RABBITMQ_SENSOR_QUEUE")
	if sensorQueueName == "" {
		sensorQueueName = "sensors.gates"
	}

	telemetryQueueName := os.Getenv("RABBITMQ_TELEMETRY_QUEUE")
	if telemetryQueueName == "" {
		telemetryQueueName = "telemetry" // значение по умолчанию
	}

	// Подключаемся к БД
	database, err := db.New(databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Создаем RabbitMQ publisher
	publisher, err := rabbitmq.NewPublisher(rabbitmqURL, telemetryQueueName)
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ publisher: %v", err)
	}
	defer publisher.Close()

	// Consumer для событий новых сенсоров
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sensorConsumer, err := rabbitmq.NewConsumer(rabbitmqURL, sensorQueueName, database)
	if err != nil {
		log.Fatalf("Failed to create sensor consumer: %v", err)
	}
	defer sensorConsumer.Close()

	go func() {
		if err := sensorConsumer.Start(ctx); err != nil && err != context.Canceled {
			log.Printf("Sensor consumer error: %v", err)
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

	// REST API для обновления статуса ворот
	r.POST("/gates", func(c *gin.Context) {
		var gatesUpdate models.GatesUpdate
		if err := c.ShouldBindJSON(&gatesUpdate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
				"details": err.Error(),
			})
			return
		}

		// Обновляем запись в БД
		gates, err := database.UpdateGates(c.Request.Context(), gatesUpdate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update gates",
				"details": err.Error(),
			})
			return
		}

		// Формируем JSON для telemetry сервиса
		telemetryValue := map[string]interface{}{
			"status": gatesUpdate.Status,
			"type":   "gates",
		}
		valueJSON, err := json.Marshal(telemetryValue)
		if err != nil {
			log.Printf("Failed to marshal telemetry value: %v", err)
			// Продолжаем выполнение, даже если не удалось отправить в RabbitMQ
		} else {
			// Отправляем сообщение в RabbitMQ для telemetry сервиса
			if err := publisher.PublishTelemetry(gatesUpdate.SerialNumber, valueJSON); err != nil {
				log.Printf("Failed to publish telemetry message: %v", err)
				// Продолжаем выполнение, даже если не удалось отправить в RabbitMQ
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Gates updated successfully",
			"data":   gates,
		})
	})

	// REST API для получения статуса по serial_number
	r.GET("/gates/:serial_number", func(c *gin.Context) {
		serialNumber := c.Param("serial_number")

		gates, err := database.GetGates(c.Request.Context(), serialNumber)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Gates not found",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gates,
		})
	})

	// Запускаем HTTP сервер
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- r.Run(":" + port)
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		log.Printf("Shutting down due to signal: %v", sig)
		cancel()
	case err := <-serverErr:
		if err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}
}
