package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
)

type StreamType struct {
	Sid       int    `db:"sid" json:"sid"`
	Name      string `db:"name" json:"name"`
	Des       string `db:"des" json:"des"`
	Tags      string `db:"tags" json:"tags"`
	Pages     string `db:"pages" json:"pages"`
	Type      int8   `db:"type" json:"type"`
	Count     int    `db:"count" json:"count"`
	Time      int64  `db:"time" json:"time"`
	UpTime    int64  `db:"up_time" json:"up_time"`
	Views     int    `db:"views" json:"views"`
	Favorites int    `db:"favorites" json:"favorites"`
}

type StreamLiteType struct {
	Sid   int    `db:"sid" json:"sid"`
	Name  string `db:"name" json:"name"`
	Count int    `db:"count" json:"count"`
}

var Db *sqlx.DB

func initDatabase() {
	database, err := sqlx.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4",
		Config.DbUser,
		Config.DbPass,
		Config.DbHost,
		Config.DbPort,
		Config.DbName,
	))
	if err != nil {
		log.Println("MySQL Connect Error:", err)
		return
	}
	database.SetMaxOpenConns(3000)
	database.SetMaxIdleConns(500)

	Db = database
}
