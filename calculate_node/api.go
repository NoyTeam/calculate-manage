package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"log"
	"strconv"
	"time"
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

func getApi(c *gin.Context) {
	session := sessions.Default(c)
	info := session.Get("status")
	if info != nil {
		pageInt := 1
		page := c.PostForm("page")
		if page != "" {
			pageInt, _ = strconv.Atoi(page)
		}

		data, err := queryData(pageInt)
		if err != nil {
			log.Println("queryData error:", err)
			c.JSON(200, gin.H{"status": 500})
			return
		}

		c.JSON(200, gin.H{"status": 200, "data": data})
	} else {
		c.JSON(200, gin.H{"status": 403})
	}
}

func searchApi(c *gin.Context) {
	session := sessions.Default(c)
	info := session.Get("status")
	if info != nil {
		search := c.PostForm("search")
		if search != "" {
			data, err := searchData(search)
			if err != nil {
				log.Println("searchData error:", err)
				c.JSON(200, gin.H{"status": 500})
				return
			}

			c.JSON(200, gin.H{"status": 200, "data": data})
		} else {
			c.JSON(200, gin.H{"status": 404})
		}
	} else {
		c.JSON(200, gin.H{"status": 403})
	}
}

func uploadApi(c *gin.Context) {
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
			c.JSON(200, gin.H{"status": 500})
			return
		}
		if err := c.SaveUploadedFile(file, "runtime/"+file.Filename); err != nil {
			c.JSON(200, gin.H{"status": 500})
			return
		}

		c.JSON(200, gin.H{"status": 200})
	} else {
		c.JSON(200, gin.H{"status": 403})
	}
}

func systemInfoApi(c *gin.Context) {
	session := sessions.Default(c)
	info := session.Get("status")
	if info != nil {
		memInfo, _ := mem.VirtualMemory()
		cpuPercent, _ := cpu.Percent(time.Second, false)

		c.JSON(200, gin.H{"status": 200, "cpu": cpuPercent[0], "mem": gin.H{
			"total":   memInfo.Total,
			"used":    memInfo.Used,
			"percent": memInfo.UsedPercent,
		}})
	} else {
		c.JSON(200, gin.H{"status": 403})
	}
}
