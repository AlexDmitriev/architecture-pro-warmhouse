package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"heating/db"
	"heating/models"
	"heating/rabbitmq"

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

	// Настраиваем HTTP сервер
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// REST API для обновления температуры
	r.POST("/heating", func(c *gin.Context) {
		var heatingUpdate models.HeatingUpdate
		if err := c.ShouldBindJSON(&heatingUpdate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
				"details": err.Error(),
			})
			return
		}

		// Обновляем запись в БД
		heating, err := database.UpdateHeating(c.Request.Context(), heatingUpdate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update heating",
				"details": err.Error(),
			})
			return
		}

		// Формируем JSON для telemetry сервиса
		telemetryValue := map[string]interface{}{
			"temperature": heatingUpdate.Temperature,
			"type":        "heating",
		}
		valueJSON, err := json.Marshal(telemetryValue)
		if err != nil {
			log.Printf("Failed to marshal telemetry value: %v", err)
			// Продолжаем выполнение, даже если не удалось отправить в RabbitMQ
		} else {
			// Отправляем сообщение в RabbitMQ для telemetry сервиса
			if err := publisher.PublishTelemetry(heatingUpdate.SerialNumber, valueJSON); err != nil {
				log.Printf("Failed to publish telemetry message: %v", err)
				// Продолжаем выполнение, даже если не удалось отправить в RabbitMQ
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Heating updated successfully",
			"data":    heating,
		})
	})

	// REST API для получения температуры по serial_number
	r.GET("/heating/:serial_number", func(c *gin.Context) {
		serialNumber := c.Param("serial_number")

		heating, err := database.GetHeating(c.Request.Context(), serialNumber)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Heating not found",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": heating,
		})
	})

	// Запускаем HTTP сервер
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	if err := r.Run(":" + port); err != nil {
		log.Printf("HTTP server error: %v", err)
	}
}
