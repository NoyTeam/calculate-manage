package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type StreamType struct {
	Sid       int    `json:"sid"`
	Name      string `json:"name"`
	Des       string `json:"des"`
	Tags      string `json:"tags"`
	Pages     string `json:"pages"`
	Type      int8   `json:"type"`
	Count     int    `json:"count"`
	Time      int64  `json:"time"`
	UpTime    int64  `json:"up_time"`
	Views     int    `json:"views"`
	Favorites int    `json:"favorites"`
}

type ApiType struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
}

var nvEnabled = false
var progressRegexp = regexp.MustCompile(`time=(.*?)bitrate=`)

func GetSHA256HashCode(message []byte) string {
	hash := sha256.New()
	hash.Write(message)
	return hex.EncodeToString(hash.Sum(nil))
}

func ffmpeg(inputFile, outputFile string, taskId string) error {
	var cmd *exec.Cmd
	if nvEnabled {
		cmd = exec.Command("ffmpeg", "-i", inputFile, "-c:v", "hevc_nvenc", "-crf", "23", "-progress", "pipe:1", outputFile)
	} else {
		cmd = exec.Command("ffmpeg", "-i", inputFile, "-c:v", "libx265", "-crf", "23", "-progress", "pipe:1", outputFile)
	}
	stderr, _ := cmd.StderrPipe()
	scanner := bufio.NewScanner(stderr)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexByte(data, '\r'); i >= 0 {
			return i + 1, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return
	})

	totalLength := getTotalLength(inputFile)

	statusId := setStatus(taskId, "Encoder is running", 0)

	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			match := progressRegexp.FindStringSubmatch(line)
			if len(match) == 2 {
				timeParts := strings.Split(strings.TrimSpace(match[1]), ":")
				if len(timeParts) == 3 {
					hours, _ := strconv.ParseFloat(timeParts[0], 64)
					minutes, _ := strconv.ParseFloat(timeParts[1], 64)
					seconds, _ := strconv.ParseFloat(timeParts[2], 64)
					currentSeconds := hours*3600 + minutes*60 + seconds
					progress := (currentSeconds / totalLength) * 100

					setStatusItem(taskId, statusId, int(progress))
				}
			}
		}
	}()

	err := cmd.Run()
	if err == nil {
		setStatusItem(taskId, statusId, 100)
	}
	return err
}

func getTotalLength(inputFile string) float64 {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", inputFile)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("ffprobe error: %v", err)
	}
	totalLength, _ := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	return totalLength
}

func PostFormRequest(url string, data url.Values) ([]byte, error) {
	resp, err := http.PostForm(url, data)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("Error closing body:", err)
		}
	}(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

func c2webp(input, output string) error {
	cmd := exec.Command("magick", input, output)
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
