package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

var Db *sql.DB

type Data struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func initDatabase() {
	var err error
	Db, err = sql.Open("sqlite3", Config.Database)
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	if err = Db.Ping(); err != nil {
		log.Fatalln("Error connecting to database:", err)
		return
	}
}

func queryStreamData(page int) ([]Data, error) {
	rows, err := Db.Query("SELECT id, name, count FROM stream ORDER BY id DESC LIMIT ?,20", (page-1)*20)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Panicln("Error closing rows:", err)
		}
	}(rows)

	var result = make([]Data, 0)
	for rows.Next() {
		var d Data
		if err := rows.Scan(&d.ID, &d.Name, &d.Count); err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func queryStreamCount() (int, error) {
	var count = 0
	err := Db.QueryRow("SELECT COUNT(id) FROM stream").Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func searchStreamData(search string, page int) ([]Data, int, error) {
	rows, err := Db.Query("SELECT id, name, count FROM stream WHERE name LIKE ? ORDER BY id DESC LIMIT ?,20", "%"+search+"%", (page-1)*20)
	if err != nil {
		return nil, 0, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Panicln("Error closing rows:", err)
		}
	}(rows)

	var result = make([]Data, 0)
	for rows.Next() {
		var d Data
		if err := rows.Scan(&d.ID, &d.Name, &d.Count); err != nil {
			return nil, 0, err
		}
		result = append(result, d)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var count = 0
	err = Db.QueryRow("SELECT COUNT(id) FROM stream WHERE name LIKE ?", "%"+search+"%").Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	return result, count, nil
}

func userExists(username, password string) (bool, error) {
	var count int
	err := Db.QueryRow("SELECT COUNT(id) FROM user WHERE username = ? AND password = ?", username, password).Scan(&count)
	if err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	return false, nil
}
