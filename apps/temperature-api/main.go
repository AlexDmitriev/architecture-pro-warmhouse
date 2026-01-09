package main

import (
    "math"
    "math/rand"
	"time"
	"net/http"
	"github.com/gin-gonic/gin"
)

type TemperatureData struct {
	Value       float64   `json:"value"`
	Unit        string    `json:"unit"`
	Timestamp   time.Time `json:"timestamp"`
	Location    string    `json:"location"`
	Status      string    `json:"status"`
	SensorID    string    `json:"sensor_id"`
	SensorType  string    `json:"sensor_type"`
	Description string    `json:"description"`
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func randomTemperature() float64 {
	value := 5.0 + rng.Float64()*(25.0-5.0)
	return math.Round(value*100) / 100
}

func main() {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": "ok",
        })
    })

    r.GET("/temperature", func(c *gin.Context) {
        location := c.Query("location")
        if location == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Location is required"})
            return
        }

        data := generateTemperatureData(location, "")

        c.JSON(http.StatusOK, data)
    })


	r.GET("/temperature/:id", func(c *gin.Context) {
	    sensorID := c.Param("id")
	    if sensorID == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Sensor ID is required"})
            return
        }

		data := generateTemperatureData("", sensorID)

		c.JSON(http.StatusOK, data)
	})

	// Слушаем порт 8081 внутри контейнера
	r.Run(":8081")
}

func generateTemperatureData(location, sensorID string) TemperatureData {
    if location == "" {
        switch sensorID {
        case "1":
            location = "Living Room"
        case "2":
            location = "Bedroom"
        case "3":
            location = "Kitchen"
        default:
            location = "Unknown"
        }
    }

    if sensorID == "" {
        switch location {
            case "Living Room":
                sensorID = "1"
            case "Bedroom":
                sensorID = "2"
            case "Kitchen":
                sensorID = "3"
            default:
                sensorID = "0"
        }
    }

    return TemperatureData{
        Value:       randomTemperature(),
        Unit:        "°C",
        Timestamp:   time.Now().UTC(),
        Location:    location,
        Status:      "active",
        SensorID:    sensorID,
        SensorType:  "temperature",
        Description: "Temperature sensor in " + location,
    }
}