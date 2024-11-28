package models

import (
	"fmt"
	"time"

	"github.com/andrew-lawlor/librepub/database"
)

type Annotation struct {
	ID             int
	EReaderID      string
	BookID         string
	Text           string
	Note           string
	AnnotationType string
	UserID         int
	Created        string
}

func CreateAnnotation(eReaderID string, bookID int, text string, note string, annotationType string, userID int) (int, error) {
	db := database.GetDB()
	var created = time.Now().Format(time.RFC3339)
	statement, err := db.Prepare("INSERT INTO annotations (text, note, type, created, ereader_id, book_id, user_id) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return -1, err
	}
	res, err := statement.Exec(text, note, annotationType, created, eReaderID, bookID, userID)
	if err != nil {
		return -1, err
	}
	annotID, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}
	return int(annotID), err
}

func GetAnnotations(userID int, bookID int) ([]Annotation, error) {
	var annotations = []Annotation{}
	db := database.GetDB()
	stmt, err := db.Prepare("SELECT id, text, note, type, created FROM annotations WHERE user_id = ? AND book_id = ?")
	if err != nil {
		WriteLog(LogError, err.Error())
		return annotations, err
	}
	row, err := stmt.Query(userID, bookID)
	if err != nil {
		WriteLog(LogError, err.Error())
		return annotations, err
	}
	defer row.Close()
	for row.Next() { // Iterate and fetch the records from result cursor
		var annotation Annotation
		err = row.Scan(&annotation.ID, &annotation.Text, &annotation.Note, &annotation.AnnotationType, &annotation.Created)
		if err != nil {
			WriteLog(LogError, err.Error())
			return annotations, err
		}
		annotations = append(annotations, annotation)
	}
	return annotations, nil
}

func AnnotationExists(eReaderID string, userID int) bool {
	db := database.GetDB()
	stmt, err := db.Prepare("SELECT id FROM annotations WHERE user_id = ? AND ereader_id = ?")
	if err != nil {
		WriteLog(LogError, err.Error())
		return false
	}
	var annotation Annotation
	err = stmt.QueryRow(userID, eReaderID).Scan(&annotation.ID)
	if err != nil {
		WriteLog(LogError, err.Error())
		return false
	}
	fmt.Println(annotation.ID)
	return true
}
