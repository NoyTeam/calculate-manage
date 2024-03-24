package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"time"
)

type StatusType struct {
	Msg     string `json:"msg"`
	Percent int    `json:"percent"`
}

var Status = make(map[string][]StatusType)

func setStatus(key string, msg string, percent int) int {
	if _, ok := Status[key]; ok {
		Status[key] = append(Status[key], StatusType{msg, percent})
		return len(Status[key]) - 1
	} else {
		Status[key] = []StatusType{{msg, percent}}
		return 0
	}
}

func setStatusItem(key string, index int, percent int) {
	if _, ok := Status[key]; ok {
		if index < len(Status[key]) {
			Status[key][index].Percent = percent
		}
	}
}

func removeStatus(key string) {
	go func() {
		time.Sleep(10 * time.Second)
		delete(Status, key)
	}()
}

func getStatusApi(c *gin.Context) {
	session := sessions.Default(c)
	info := session.Get("status")
	if info != nil {
		key := c.Query("key")

		if _, ok := Status[key]; ok {
			c.JSON(200, gin.H{"status": 200, "data": Status[key]})
		} else {
			c.JSON(200, gin.H{"status": 404})
		}
	}
}
