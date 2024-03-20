package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"log"
)

func loginApi(c *gin.Context) {
	user := c.PostForm("username")
	pass := GetSHA256HashCode([]byte(c.PostForm("password") + Config.Secret))

	session := sessions.Default(c)
	exists, err := userExists(user, pass)
	if err != nil {
		log.Println("userExists error:", err)
		c.JSON(200, gin.H{"status": 500})
		return
	}
	if exists {
		session.Set("status", "ok")
		err := session.Save()
		if err != nil {
			c.JSON(200, gin.H{"status": 500})
			return
		}

		c.JSON(200, gin.H{"status": 200})
	} else {
		c.JSON(200, gin.H{"status": 403})
	}
}

func loginOut(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("status") != nil {
		session.Clear()
		session.Options(sessions.Options{Path: "/", MaxAge: -1})
		err := session.Save()
		if err != nil {
			c.JSON(200, gin.H{"status": 500})
			return
		}
		c.JSON(200, gin.H{"status": 200})
	} else {
		c.JSON(200, gin.H{"status": 403})
	}
}
