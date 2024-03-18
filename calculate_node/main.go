package main

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

func StaticFileHandler(c *gin.Context) {
	filePath := c.Request.URL.Path
	if filePath == "/" {
		filePath = "/index"
	}

	switch filepath.Ext(filePath) {
	case ".js":
		c.Header("Content-Type", "application/javascript")
	case ".css":
		c.Header("Content-Type", "text/css")
	case ".webp":
		c.Header("Content-Type", "image/webp")
	case ".woff2":
		c.Header("Content-Type", "font/woff2")
	case ".woff":
		c.Header("Content-Type", "font/woff")
	case ".ttf":
		c.Header("Content-Type", "font/ttf")
	default:
		c.Header("Content-Type", "text/html")
		filePath = filePath + ".html"
	}

	fullPath := "./web" + filePath
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"status": 404, "error": "Not Found"})
		return
	}

	c.File(fullPath)
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	store := cookie.NewStore([]byte(Config.Secret))
	router.Use(sessions.Sessions("SESSION", store))
	router.Use(cors())
	router.Use(gin.Recovery())

	// 404
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"status": 404})
	})

	// API
	router.POST("/api/login", loginApi) // Login

	router.POST("/api/get", getApi)       // Get Stream
	router.POST("/api/search", searchApi) // Search Stream
	router.POST("/api/upload", uploadApi) // Upload Stream

	router.GET("/api/system/info", systemInfoApi) // System Info

	// Web
	router.NoRoute(StaticFileHandler)

	// Run
	host := fmt.Sprintf("%s:%d", Config.Host, Config.Port)
	log.Printf("Server Starting: %s ...\n", host)
	err := router.Run(host)
	if err != nil {
		log.Println("Server Start Failed ...")
	}
}
