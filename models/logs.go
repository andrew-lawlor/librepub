package models

import (
	"log"
	"strconv"
	"time"

	"github.com/andrew-lawlor/librepub/database"
)

const (
	LogInfo  = "Info"
	LogDebug = "Debug"
	LogError = "Error"

	page_offset = 5
)

type Log struct {
	ID      int
	Level   string
	Created string
	Message string
}

func NewLog(id int, level, created, message string) Log {
	return Log{
		ID:      id,
		Level:   level,
		Created: created,
		Message: message,
	}
}

func WriteLog(level string, message string) bool {
	db := database.GetDB()
	statement, err := db.Prepare("INSERT INTO logs (level, created, message) VALUES (?, ?, ?)")
	if err != nil {
		return false
	}
	var created = time.Now().Format(time.RFC3339)
	_, err = statement.Exec(level, created, message)
	return err == nil
}

func GetLogs(page int, search string) []Log {
	var offset = (page - 1) * page_offset
	db := database.GetDB()
	var query string
	var queryArg string
	if search == "" {
		query = "SELECT * FROM logs ORDER BY created DESC LIMIT 5 OFFSET ?"
		queryArg = strconv.Itoa(offset)
	} else {
		query = "SELECT * FROM logs WHERE message LIKE ? ORDER BY created DESC LIMIT 5 OFFSET " + strconv.Itoa(offset)
		queryArg = "%" + search + "%"
	}
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}
	row, err := stmt.Query(queryArg)
	if err != nil {
		log.Fatal(err)
	}
	var logs = []Log{}
	defer row.Close()
	for row.Next() { // Iterate and fetch the records from result cursor
		var log Log
		err = row.Scan(&log.ID, &log.Level, &log.Created, &log.Message)
		if err != nil {
			WriteLog(LogError, err.Error())
		}
		logs = append(logs, log)
	}
	return logs
}

func DeleteLog(id int) bool {
	db := database.GetDB()
	statement, _ := db.Prepare("DELETE from logs WHERE id = ?")
	_, err := statement.Exec(id)
	return err == nil
}

func DeleteAllLogs() bool {
	db := database.GetDB()
	statement, _ := db.Prepare("DELETE from logs")
	_, err := statement.Exec()
	return err == nil
}
