package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"path"
	"strconv"
	"strings"
	"time"
)

func getStreamInfo(c *gin.Context) {
	token := c.PostForm("token")
	if token == Config.Token {
		sid := c.PostForm("sid")
		sidInt, _ := strconv.Atoi(sid)

		var data = make([]StreamType, 0)
		err := Db.Select(&data, "SELECT sid, name, des, tags, pages, type, count, time, up_time, views, favorites FROM stream WHERE `sid` = ?", sidInt)
		if err != nil {
			c.JSON(200, gin.H{"status": 500, "msg": err.Error()})
			return
		}

		if len(data) == 1 {
			c.JSON(200, gin.H{"status": 200, "data": data[0]})
		} else {
			c.JSON(200, gin.H{"status": 404})
		}
	} else {
		c.JSON(200, gin.H{"status": 403})
	}
}

func getStreamList(c *gin.Context) {
	token := c.PostForm("token")
	if token == Config.Token {
		var data = make([]StreamLiteType, 0)

		err := Db.Select(&data, "SELECT sid, name, count FROM stream")
		if err != nil {
			c.JSON(200, gin.H{"status": 500, "msg": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": 200, "data": data})
	} else {
		c.JSON(200, gin.H{"status": 403})
	}
}

func uploadStream(c *gin.Context) {
	token := c.PostForm("token")
	if token == Config.Token {
		sid := c.PostForm("sid")
		sidInt, _ := strconv.Atoi(sid)
		pid := c.PostForm("pid")
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(200, gin.H{"status": 500, "msg": err.Error()})
			return
		}

		var data = make([]StreamType, 0)
		err = Db.Select(&data, "SELECT sid, name, des, tags, pages, type, count, time, up_time, views, favorites FROM stream WHERE `sid` = ?", sidInt)
		if err != nil {
			c.JSON(200, gin.H{"status": 500, "msg": err.Error()})
			return
		}

		if len(data) == 1 {
			pidInt, _ := strconv.Atoi(pid)
			if data[0].Count < pidInt {
				newFileName := GetSHA256HashCode([]byte(sid+FileToken+pid)) + "-" + pid + ".mp4"
				if err := c.SaveUploadedFile(file, path.Join("stream", sid, newFileName)); err != nil {
					log.Println("SaveUploadedFile error:", err)
					c.JSON(200, gin.H{"status": 500, "msg": "SaveUploadedFile error: " + err.Error()})
					return
				}

				var pages = make([]string, 0, data[0].Count+1)
				for i := 1; i <= data[0].Count+1; i++ {
					pages = append(pages, strconv.Itoa(i))
				}

				_, err = Db.Exec("UPDATE stream SET count = count + 1, pages = ? WHERE `sid` = ?", strings.Join(pages, ","), sidInt)
				if err != nil {
					c.JSON(200, gin.H{"status": 500, "msg": err.Error()})
					return
				}

				c.JSON(200, gin.H{"status": 200})
			} else {
				c.JSON(200, gin.H{"status": 500, "msg": "pid error (" + pid + ")"})
			}
		} else {
			c.JSON(200, gin.H{"status": 404})
		}
	} else {
		c.JSON(200, gin.H{"status": 403})
	}
}

func createStream(c *gin.Context) {
	token := c.PostForm("token")
	if token == Config.Token {
		name := c.PostForm("name")
		des := c.PostForm("des")
		tags := c.PostForm("tags")
		typeStr := c.PostForm("type")
		now := time.Now().Unix()

		var data = make([]StreamType, 0)
		err := Db.Select(&data, "SELECT sid FROM stream WHERE `name` = ?", name)
		if err != nil {
			c.JSON(200, gin.H{"status": 500, "msg": err.Error()})
			return
		}

		if len(data) == 0 {
			_, err = Db.Exec("INSERT INTO stream (name, des, tags, pages, type, count, time, up_time, views, faveorites) VALUES (?, ?, ?, \"\", ?, 0, ?, ?, 0, 0)", name, des, tags, typeStr, now, now)
			if err != nil {
				c.JSON(200, gin.H{"status": 500, "msg": err.Error()})
				return
			}

			c.JSON(200, gin.H{"status": 200})
		} else {
			c.JSON(200, gin.H{"status": 500, "msg": "name exists"})
		}
	} else {
		c.JSON(200, gin.H{"status": 403})
	}
}
