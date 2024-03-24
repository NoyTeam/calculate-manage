package main

import (
	"encoding/json"
	"log"
	"os"
)

var Config struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Token    string `json:"token"`
	Server   string `json:"server"`
	Secret   string `json:"secret"`
	Database string `json:"database"`
}

func init() {
	// Config
	file, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalln("config.json not found")
	}
	err = json.Unmarshal(file, &Config)
	if err != nil {
		log.Fatalln("config.json is not valid:", err)
	}

	// Runtime Path
	_, err = os.Stat("runtime")
	if os.IsNotExist(err) {
		err = os.Mkdir("runtime", 0755)
		if err != nil {
			log.Fatalln("runtime directory creation failed")
		}
	}

	// Database
	initDatabase()
}
