package main

import (
	"encoding/json"
	"log"
	"os"
)

var Config struct {
	Host   string `json:"host"`
	Port   int    `json:"port"`
	Token  string `json:"token"`
	DbHost string `json:"db_host"`
	DbPort int    `json:"db_port"`
	DbUser string `json:"db_user"`
	DbPass string `json:"db_pass"`
	DbName string `json:"db_name"`
}

func init() {
	file, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalln("config.json not found")
	}
	err = json.Unmarshal(file, &Config)
	if err != nil {
		log.Fatalln("config.json is not valid:", err)
	}

	// Init Database
	initDatabase()
}
