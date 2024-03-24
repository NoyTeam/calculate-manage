package main

import (
	"embed"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"log"
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

//go:embed web
var webFS embed.FS

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

	//fullPath := "./web" + filePath
	//if _, err := os.Stat(fullPath); os.IsNotExist(err) {
	//	c.JSON(404, gin.H{"status": 404, "error": "Not Found"})
	//	return
	//}
	//c.File(fullPath)

	file, err := webFS.ReadFile("web" + filePath)
	if err != nil {
		c.JSON(404, gin.H{"status": 404, "error": "Not Found"})
		return
	}
	c.String(200, string(file))
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	store := cookie.NewStore([]byte(Config.Secret))
	router.Use(sessions.Sessions("SESSION", store))
	router.Use(cors())
	router.Use(gin.Recovery())

	// API
	router.POST("/api/login", loginApi) // Login
	router.GET("/api/logout", loginOut) // Logout

	router.POST("/api/stream/get", getStreamApi)       // Get Stream
	router.POST("/api/stream/sync", streamSyncApi)     // Sync Stream
	router.POST("/api/stream/search", searchStreamApi) // Search Stream
	router.POST("/api/stream/upload", uploadStreamApi) // Upload Stream

	router.GET("/api/system/info", systemInfoApi)        // System Info
	router.POST("/api/system/command", systemCommandApi) // System Command
	router.GET("/api/system/init", getInitInfoApi)       // System Init Info
	router.POST("/api/system/fan", fanSetApi)            // Fan Speed

	router.GET("/api/status/get", getStatusApi) // Get Status

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
