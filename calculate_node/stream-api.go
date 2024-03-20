package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"log"
	"strconv"
)

func getStreamApi(c *gin.Context) {
	session := sessions.Default(c)
	info := session.Get("status")
	if info != nil {
		pageInt := 1
		page := c.PostForm("page")
		if page != "" {
			pageInt, _ = strconv.Atoi(page)
		}

		data, err := queryStreamData(pageInt)
		if err != nil {
			log.Println("queryStreamData error:", err)
			c.JSON(200, gin.H{"status": 500, "msg": err.Error()})
			return
		}

		count, err := queryStreamCount()
		if err != nil {
			log.Println("queryStreamCount error:", err)
			c.JSON(200, gin.H{"status": 500, "msg": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": 200, "data": data, "count": count})
	} else {
		c.JSON(200, gin.H{"status": 403})
	}
}

func searchStreamApi(c *gin.Context) {
	session := sessions.Default(c)
	info := session.Get("status")
	if info != nil {
		pageInt := 1
		page := c.PostForm("page")
		if page != "" {
			pageInt, _ = strconv.Atoi(page)
		}

		search := c.PostForm("search")
		if search != "" {
			data, count, err := searchStreamData(search, pageInt)
			if err != nil {
				log.Println("searchStreamData error:", err)
				c.JSON(200, gin.H{"status": 500, "msg": err.Error()})
				return
			}

			c.JSON(200, gin.H{"status": 200, "data": data, "count": count})
		} else {
			c.JSON(200, gin.H{"status": 404})
		}
	} else {
		c.JSON(200, gin.H{"status": 403})
	}
}

func uploadStreamApi(c *gin.Context) {
	session := sessions.Default(c)
	info := session.Get("status")
	if info != nil {
		id, _ := strconv.Atoi(c.PostForm("id"))
		if id == 0 {
			c.JSON(200, gin.H{"status": 404})
			return
		}

		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(200, gin.H{"status": 500, "msg": err.Error()})
			return
		}
		if err := c.SaveUploadedFile(file, "runtime/"+file.Filename); err != nil {
			c.JSON(200, gin.H{"status": 500, "msg": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": 200})
	} else {
		c.JSON(200, gin.H{"status": 403})
	}
}
