package main

import (
	"encoding/json"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"
)

type GetStreamApiType struct {
	Status int        `json:"status"`
	Data   StreamType `json:"data"`
}

type GetStreamListApiType struct {
	Status int          `json:"status"`
	Data   []StreamType `json:"data"`
}

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

func streamSyncApi(c *gin.Context) {
	session := sessions.Default(c)
	info := session.Get("status")
	if info != nil {
		upStreamAll()
		c.JSON(200, gin.H{"status": 200})
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
			c.JSON(200, gin.H{"status": 404, "msg": "stream id not found"})
			return
		}
		file, err := c.FormFile("file")
		if err != nil {
			log.Println("FormFile error:", err)
			c.JSON(200, gin.H{"status": 500, "msg": "FormFile error: " + err.Error()})
			return
		}
		runtimeFilePath := "runtime/" + file.Filename
		if err := c.SaveUploadedFile(file, runtimeFilePath); err != nil {
			log.Println("SaveUploadedFile error:", err)
			c.JSON(200, gin.H{"status": 500, "msg": "SaveUploadedFile error: " + err.Error()})
			return
		}

		var taskId = strconv.FormatInt(time.Now().Unix(), 10)

		setStatus(taskId, "Start Task", -1)

		go func() {
			// Get Info
			request, err := PostFormRequest(Config.Server+"/api/stream/get", url.Values{"sid": {strconv.Itoa(id)}, "token": {Config.Token}})
			if err != nil {
				log.Println("PostFormRequest error:", err)
				setStatus(taskId, "Error: "+err.Error(), -1)
				return
			}
			var streamData GetStreamApiType
			if err := json.Unmarshal(request, &streamData); err != nil {
				log.Println("json.Unmarshal error:", err)
				setStatus(taskId, "Error: "+err.Error(), -1)
				return
			}

			if streamData.Status == 200 {
				// FFMpeg
				setStatus(taskId, "Encoder starting", -1)
				outputFilePath := "runtime/" + strconv.Itoa(id) + "-" + taskId + ".mp4"
				if err := ffmpeg(runtimeFilePath, outputFilePath, taskId); err != nil {
					log.Println("ffmpeg error:", err)
					setStatus(taskId, "Error: "+err.Error(), -1)
					return
				}

				defer func() {
					if err := os.Remove(outputFilePath); err != nil {
						log.Println("Remove output file error:", err)
					}
					if err := os.Remove(runtimeFilePath); err != nil {
						log.Println("Remove runtime file error:", err)
					}
				}()

				// Upload
				setStatus(taskId, "Upload starting", -1)
				if err := uploadStreamFile(outputFilePath, strconv.Itoa(id), strconv.Itoa(streamData.Data.Count+1), taskId); err != nil {
					log.Println("uploadStreamFile error:", err)
					setStatus(taskId, "Error: "+err.Error(), -1)
					return
				}

				upStreamDbOne(id)

				setStatus(taskId, "Stream Upload success: "+taskId, -1)
			} else {
				setStatus(taskId, "Stream Upload failed: "+taskId+", Error: "+strconv.Itoa(streamData.Status), -1)
			}

			removeStatus(taskId)
		}()

		c.JSON(200, gin.H{"status": 200, "task_id": taskId})
	} else {
		c.JSON(200, gin.H{"status": 403})
	}
}
