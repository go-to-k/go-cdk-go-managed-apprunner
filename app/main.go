package main

import (
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	outputValue := os.Getenv("ENV1")
	engine := gin.Default()
	engine.GET("/", func(c *gin.Context) {
		c.String(200, outputValue)
	})
	engine.Run(":8080")
}
