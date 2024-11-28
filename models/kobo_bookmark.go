package models

import (
	"database/sql"
)

type KoboBookmark struct {
	BookmarkID     string
	VolumeID       string
	Text           string
	Annotation     string
	AnnotationType string
	BookTitle      string
}

func NewKoboBookmark(id string, volume string, txt string, note string, noteType string) KoboBookmark {
	return KoboBookmark{
		BookmarkID:     id,
		VolumeID:       volume,
		Text:           txt,
		Annotation:     note,
		AnnotationType: noteType,
	}
}

func GetKoboBookmarks(dbPath string) ([]KoboBookmark, error) {
	var bookmarks = []KoboBookmark{}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		WriteLog(LogError, err.Error())
		return bookmarks, err
	}
	stmt, err := db.Prepare("SELECT BookmarkID, VolumeID, Text, Annotation, Type FROM Bookmark WHERE Type IN (?, ?)")
	if err != nil {
		WriteLog(LogError, err.Error())
		return bookmarks, err
	}
	row, err := stmt.Query("highlight", "note")
	if err != nil {
		WriteLog(LogError, err.Error())
		return bookmarks, err
	}
	defer row.Close()
	for row.Next() { // Iterate and fetch the records from result cursor
		var bookmark KoboBookmark
		err = row.Scan(&bookmark.BookmarkID, &bookmark.VolumeID, &bookmark.Text, &bookmark.Annotation, &bookmark.AnnotationType)
		if err != nil {
			WriteLog(LogError, err.Error())
			return bookmarks, err
		}
		bookmark.BookTitle = GetBookTitle(dbPath, bookmark.VolumeID)
		bookmarks = append(bookmarks, bookmark)
	}
	return bookmarks, nil
}

func GetBookTitle(dbPath string, volumeID string) string {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		WriteLog(LogError, err.Error())
		return ""
	}
	stmt, err := db.Prepare("SELECT BookTitle FROM content WHERE BookID = ?")
	if err != nil {
		WriteLog(LogError, err.Error())
		return ""
	}
	var bookTitle string
	err = stmt.QueryRow(volumeID).Scan(&bookTitle)
	if err != nil {
		WriteLog(LogError, err.Error())
		return ""
	}
	return bookTitle
}
