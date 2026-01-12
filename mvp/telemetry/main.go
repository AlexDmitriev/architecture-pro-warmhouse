package main

import (
    "time"
	"net/http"
	"github.com/gin-gonic/gin"
)

type TelemetryData struct {
    Status      string    `json:"status"`
	SensorID    string    `json:"sensor_id"`
	Value       string    `json:"value"`
	Timestamp   time.Time `json:"timestamp"`
}

func main() {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": "ok",
        })
    })

	r.Run(":8081")
}