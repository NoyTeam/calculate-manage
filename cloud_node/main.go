package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
)

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

var FileToken = "noy_stream_24578"

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(cors())
	router.Use(gin.Recovery())

	// 404
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"status": 404})
	})

	// API
	router.POST("/api/stream/get", getStreamInfo)
	router.POST("/api/stream/list", getStreamList)
	router.POST("/api/stream/upload", uploadStream)
	router.POST("/api/stream/create", createStream)

	// Run
	host := fmt.Sprintf("%s:%d", Config.Host, Config.Port)
	log.Printf("Server Starting: %s ...\n", host)
	err := router.Run(host)
	if err != nil {
		log.Println("Server Start Failed ...")
	}
}
