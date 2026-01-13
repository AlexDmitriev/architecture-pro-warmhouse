package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"monolith/db"
	"monolith/models"
	"monolith/rabbitmq"

	"github.com/gin-gonic/gin"
)

var allowedTypes = map[string]struct{}{
	"gate":    {},
	"heating": {},
}

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		log.Fatal("RABBITMQ_URL environment variable is required")
	}

	gatesQueue := os.Getenv("RABBITMQ_GATES_QUEUE")
	if gatesQueue == "" {
		gatesQueue = "sensors.gates"
	}

	heatingQueue := os.Getenv("RABBITMQ_HEATING_QUEUE")
	if heatingQueue == "" {
		heatingQueue = "sensors.heating"
	}

	database, err := db.New(databaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	publisher, err := rabbitmq.NewPublisher(rabbitURL, map[string]string{
		"gate":    gatesQueue,
		"heating": heatingQueue,
	})
	if err != nil {
		log.Fatalf("failed to create rabbit publisher: %v", err)
	}
	defer publisher.Close()

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/sensors", func(c *gin.Context) {
		items, err := database.ListSensors(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "count": len(items)})
	})

	r.GET("/sensors/:id", func(c *gin.Context) {
		id := c.Param("id")
		sensor, err := database.GetSensor(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": sensor})
	})

	r.POST("/sensors", func(c *gin.Context) {
		var input models.SensorCreate
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if _, ok := allowedTypes[input.Type]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unknown sensor type"})
			return
		}

		sensor, err := database.CreateSensor(c.Request.Context(), input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := publisher.PublishSensorCreated(
			c.Request.Context(),
			input.Type,
			rabbitmq.SensorMessage{
				SensorID:     sensor.ID,
				SerialNumber: sensor.SerialNumber,
				Type:         input.Type,
			},
		); err != nil {
			log.Printf("failed to publish sensor message: %v", err)
		}

		c.JSON(http.StatusCreated, gin.H{"data": sensor})
	})

	r.PUT("/sensors/:id", func(c *gin.Context) {
		id := c.Param("id")
		var input models.SensorUpdate
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if input.Type != nil {
			if _, ok := allowedTypes[*input.Type]; !ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": "unknown sensor type"})
				return
			}
		}

		sensor, err := database.UpdateSensor(c.Request.Context(), id, input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": sensor})
	})

	r.DELETE("/sensors/:id", func(c *gin.Context) {
		id := c.Param("id")
		if err := database.DeleteSensor(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	// Грейсфул завершение
	srvErr := make(chan error, 1)
	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		srvErr <- r.Run(":" + port)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-srvErr:
		if err != nil {
			log.Printf("http server error: %v", err)
		}
	case sig := <-quit:
		log.Printf("shutting down due to signal: %v", sig)
	}
}

