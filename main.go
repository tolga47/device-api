package main

import (
	"log"
	"net/http"
	"os"

	"device-api/internal/device"

	"github.com/gin-gonic/gin"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "devices.db"
	}

	store, err := device.NewStore(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	device.NewHandler(store).Register(r)

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
