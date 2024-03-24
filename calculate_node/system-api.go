package main

import (
	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var nvDevice nvml.Device

func init() {
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		log.Printf("Unable to initialize NVML: %v", nvml.ErrorString(ret))
		return
	}
	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		log.Printf("Unable to get device count: %v", nvml.ErrorString(ret))
		return
	}
	if count != 0 {
		nvEnabled = true
		nvDevice, ret = nvml.DeviceGetHandleByIndex(0)
		if ret != nvml.SUCCESS {
			log.Printf("Unable to get device at index %d: %v", 0, nvml.ErrorString(ret))
		}
	}
}

func systemInfoApi(c *gin.Context) {
	session := sessions.Default(c)
	info := session.Get("status")
	if info != nil {
		memInfo, _ := mem.VirtualMemory()
		cpuPercent, _ := cpu.Percent(time.Second, false)

		var temperatureValue = "0.0"
		if runtime.GOOS == "linux" {
			// lm-sensors
			contents, err := os.ReadFile("/sys/class/hwmon/hwmon0/temp1_input")
			if err == nil {
				temperatureRaw := strings.TrimSpace(string(contents))
				temperatureValue = temperatureRaw[:2] + "." + temperatureRaw[2:]
			}
		}

		var gpuTemperature int32 = 0
		var gpuUtilization int32 = 0
		var gpuMemory int32 = 0

		if nvEnabled {
			utilization, ret := nvDevice.GetUtilizationRates()
			if ret != nvml.SUCCESS {
				log.Printf("Unable to get utilization: %v", nvml.ErrorString(ret))
			}
			temperature, ret := nvDevice.GetTemperature(nvml.TEMPERATURE_GPU)
			if ret != nvml.SUCCESS {
				log.Printf("Unable to get temperature: %v", nvml.ErrorString(ret))
			}

			gpuUtilization = int32(utilization.Gpu)
			gpuMemory = int32(utilization.Memory)
			gpuTemperature = int32(temperature)
		}

		c.JSON(200, gin.H{"status": 200, "cpu": gin.H{
			"percent": cpuPercent[0],
			"temp":    temperatureValue,
		}, "mem": gin.H{
			"total":   memInfo.Total,
			"used":    memInfo.Used,
			"percent": memInfo.UsedPercent,
		}, "gpu": gin.H{
			"temp": gpuTemperature,
			"util": gpuUtilization,
			"mem":  gpuMemory,
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
		case "help":
			msg = "Command Help:\nping - Pong!\nhelp - Show this message"
		case "ping":
			msg = "Pong!"

		default:
			msg = "Command not found"
		}
		c.JSON(200, gin.H{"status": 200, "msg": msg})
	} else {
		c.JSON(200, gin.H{"status": 403})
	}
}

func getInitInfoApi(c *gin.Context) {
	session := sessions.Default(c)
	info := session.Get("status")
	if info != nil {
		cpuInfo, err := cpu.Info()
		if err != nil {
			log.Println("cpu.Info error:", err)
			c.JSON(200, gin.H{"status": 500, "msg": err.Error()})
			return
		}
		memInfo, err := mem.VirtualMemory()
		if err != nil {
			log.Println("mem.VirtualMemory error:", err)
			c.JSON(200, gin.H{"status": 500, "msg": err.Error()})
			return
		}
		var gpuDeviceName = "N/A"
		if nvEnabled {
			name, ret := nvDevice.GetName()
			if ret != nvml.SUCCESS {
				log.Printf("Unable to get device name: %v", nvml.ErrorString(ret))
			} else {
				gpuDeviceName = name
			}
		}

		c.JSON(200, gin.H{
			"status": 200,
			"device": gin.H{
				"cpu":      cpuInfo[0].ModelName,
				"mem":      memInfo.Total,
				"gpu":      gpuDeviceName,
				"platform": runtime.GOOS,
			},
			"fans": readFanSpeed(),
		})
	} else {
		c.JSON(200, gin.H{"status": 403})
	}
}

func readFanSpeed() int {
	file, err := os.ReadFile("/etc/fancontrol")
	if err != nil {
		log.Println("Read FanControl:", err)
		return -1
	}
	lines := strings.Split(string(file), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "MINSTOP=") {
			fanSpeed := strings.Split(line, "=")[2]
			value, err := strconv.Atoi(fanSpeed)
			if err != nil {
				log.Println("Fan Speed:", err)
				return -1
			}
			return value
		}
	}

	return -1
}

func fanSetApi(c *gin.Context) {
	session := sessions.Default(c)
	info := session.Get("status")
	if info != nil {
		value := c.PostForm("value")

		file, err := os.ReadFile("/etc/fancontrol")
		if err != nil {
			log.Println("Read FanControl:", err)
			c.JSON(200, gin.H{"status": 500, "msg": "Read FanControl: " + err.Error()})
			return
		}
		lines := strings.Split(string(file), "\n")
		for i, line := range lines {
			if strings.HasPrefix(line, "MINSTOP=") {
				lineSplit := strings.Split(lines[i], "=")
				lines[i] = "MINSTOP=" + lineSplit[1] + "=" + value
				break
			}
		}
		newFile := strings.Join(lines, "\n")
		err = os.WriteFile("/etc/fancontrol", []byte(newFile), 0644)
		if err != nil {
			log.Println("Write FanControl:", err)
			c.JSON(200, gin.H{"status": 500, "msg": "Write FanControl: " + err.Error()})
			return
		}
		cmd := exec.Command("service", "fancontrol", "restart")
		err = cmd.Run()
		if err != nil {
			log.Println("Restart FanControl:", err)
			c.JSON(200, gin.H{"status": 500, "msg": "Restart FanControl: " + err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": 200})
	} else {
		c.JSON(200, gin.H{"status": 403})
	}
}
