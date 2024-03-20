package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"os"
	"runtime"
	"strings"
	"time"
)

func systemInfoApi(c *gin.Context) {
	session := sessions.Default(c)
	info := session.Get("status")
	if info != nil {
		memInfo, _ := mem.VirtualMemory()
		cpuPercent, _ := cpu.Percent(time.Second, false)

		var temperatureValue = "0.0"
		if runtime.GOOS == "linux" {
			contents, err := os.ReadFile("/sys/class/hwmon/hwmon0/temp1_input")
			if err != nil {
				temperatureRaw := strings.TrimSpace(string(contents))
				temperatureValue = temperatureRaw[:2] + "." + temperatureRaw[2:]
			}
		}

		c.JSON(200, gin.H{"status": 200, "cpu": gin.H{
			"percent": cpuPercent[0],
			"temp":    temperatureValue,
		}, "mem": gin.H{
			"total":   memInfo.Total,
			"used":    memInfo.Used,
			"percent": memInfo.UsedPercent,
		}})
	} else {
		c.JSON(200, gin.H{"status": 403})
	}
}

func systemCommandApi(c *gin.Context) {
	session := sessions.Default(c)
	info := session.Get("status")
	if info != nil {
		command := strings.Split(c.PostForm("command"), " ")

		msg := ""
		switch command[0] {
		case "pong":
			msg = "Pong!"
		default:
			msg = "Command not found"
		}
		c.JSON(200, gin.H{"status": 200, "msg": msg})
	} else {
		c.JSON(200, gin.H{"status": 403})
	}
}
