package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

type ProgressReader struct {
	Reader   io.Reader
	Progress func(int64)
	R        int64
}

func (r *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.R += int64(n)
	r.Progress(r.R)
	return n, err
}

type ProgressTransport struct {
	Transport     http.RoundTripper
	Total         int64
	UploadedBytes int64
	TaskID        string
	StatusID      int
}

func (t *ProgressTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = req.Body.Close()
		req.Body = io.NopCloser(bytes.NewReader(body))
	}()

	t.Total = int64(len(body))
	t.UploadedBytes = 0
	t.setStatus(0)

	req.Body = io.NopCloser(&ProgressReader{
		Reader:   bytes.NewReader(body),
		Progress: t.updateProgress,
	})

	return t.Transport.RoundTrip(req)
}

func (t *ProgressTransport) setStatus(progress int) {
	setStatusItem(t.TaskID, t.StatusID, progress)
}

func (t *ProgressTransport) updateProgress(uploaded int64) {
	progress := int(float64(uploaded) / float64(t.Total) * 100)
	t.setStatus(progress)
}

func uploadStreamFile(filePath, sid, pid, taskId string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println("Error closing file:", err)
		}
	}(file)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}

	_ = writer.WriteField("pid", pid)
	_ = writer.WriteField("sid", sid)
	_ = writer.WriteField("token", Config.Token)

	err = writer.Close()
	if err != nil {
		return err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	fileSize := fileInfo.Size()

	transport := &ProgressTransport{
		Transport: http.DefaultTransport,
		TaskID:    taskId,
		StatusID:  setStatus(taskId, fmt.Sprintf("Uploading (%.2f MB)", float64(fileSize)/1024/1024), 0),
	}

	client := &http.Client{
		Transport: transport,
	}

	r, err := http.NewRequest("POST", Config.Server+"/api/stream/upload", body)
	if err != nil {
		return err
	}
	r.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(r)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("Error closing body:", err)
		}
	}(resp.Body)

	apiBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var response ApiType
	err = json.Unmarshal(apiBody, &response)
	if err != nil {
		log.Println("json.Unmarshal value:", string(apiBody))
		return err
	}

	if response.Status != 200 {
		return fmt.Errorf("upload failed: %v", response)
	}

	return nil
}

func upStreamDbOne(sid int) {
	request, err := PostFormRequest(Config.Server+"/api/stream/get", url.Values{"sid": {strconv.Itoa(sid)}, "token": {Config.Token}})
	if err != nil {
		log.Println("PostFormRequest error:", err)
		return
	}
	var streamData GetStreamApiType
	if err := json.Unmarshal(request, &streamData); err != nil {
		log.Println("json.Unmarshal error:", err)
		return
	}

	if streamData.Status == 200 {
		_, err := Db.Exec("UPDATE stream SET name = ?, count = ? WHERE `id` = ?", streamData.Data.Name, streamData.Data.Count, sid)
		if err != nil {
			log.Println("Db.Exec error:", err)
			return
		}
	} else {
		log.Println("streamData.Status error:", streamData.Status)
	}
}

func upStreamAll() {
	request, err := PostFormRequest(Config.Server+"/api/stream/list", url.Values{"token": {Config.Token}})
	if err != nil {
		log.Println("PostFormRequest error:", err)
		return
	}
	var streamData GetStreamListApiType
	if err := json.Unmarshal(request, &streamData); err != nil {
		log.Println("json.Unmarshal error:", err)
		return
	}

	if streamData.Status == 200 {
		for _, v := range streamData.Data {
			_, err := Db.Exec("INSERT OR REPLACE INTO stream (id, name, count) VALUES (?, ?, ?)", v.Sid, v.Name, v.Count)
			if err != nil {
				log.Println("Db.Exec error:", err)
				return
			}
		}
	} else {
		log.Println("streamData.Status error:", streamData.Status)
	}
}
