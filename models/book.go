package models

import (
	"strconv"

	"github.com/andrew-lawlor/librepub/database"
)

type Book struct {
	ID          int
	Title       string
	EReaderID   string
	UserID      int
	Annotations []Annotation
}

func GetBooks(page int, search string, userID int) []Book {
	var offset = (page - 1) * page_offset
	db := database.GetDB()
	var query string
	var queryArg string
	if search == "" {
		query = "SELECT * FROM books WHERE user_id = ? ORDER BY title LIMIT 10 OFFSET ?"
		queryArg = strconv.Itoa(offset)
	} else {
		query = "SELECT * FROM books WHERE user_id = ? AND title LIKE ? ORDER BY title LIMIT 10 OFFSET " + strconv.Itoa(offset)
		queryArg = "%" + search + "%"
	}
	stmt, err := db.Prepare(query)
	if err != nil {
		WriteLog(LogError, err.Error())
	}
	row, err := stmt.Query(userID, queryArg)
	if err != nil {
		WriteLog(LogError, err.Error())
	}
	var books = []Book{}
	defer row.Close()
	for row.Next() { // Iterate and fetch the records from result cursor
		var book Book
		err = row.Scan(&book.ID, &book.Title, &book.EReaderID, &book.UserID)
		if err != nil {
			WriteLog(LogError, err.Error())
		}
		annotations, err := GetAnnotations(userID, book.ID)
		if err != nil {
			book.Annotations = []Annotation{}
		} else {
			book.Annotations = annotations
		}
		books = append(books, book)
	}
	return books
}

func GetAllBooks(userID int) []Book {
	db := database.GetDB()
	var query = "SELECT id, title FROM books WHERE user_id = ? ORDER BY title"
	stmt, err := db.Prepare(query)
	if err != nil {
		WriteLog(LogError, err.Error())
	}
	row, err := stmt.Query(userID)
	if err != nil {
		WriteLog(LogError, err.Error())
	}
	var books = []Book{}
	defer row.Close()
	for row.Next() { // Iterate and fetch the records from result cursor
		var book Book
		err = row.Scan(&book.ID, &book.Title)
		if err != nil {
			WriteLog(LogError, err.Error())
		}
		annotations, err := GetAnnotations(userID, book.ID)
		if err != nil {
			book.Annotations = []Annotation{}
		} else {
			book.Annotations = annotations
		}
		books = append(books, book)
	}
	return books
}

func GetBook(bookID int, userID int) Book {
	db := database.GetDB()
	stmt, err := db.Prepare("SELECT * FROM books WHERE user_id = ? AND id = ?")
	if err != nil {
		WriteLog(LogError, err.Error())
	}
	var book Book
	err = stmt.QueryRow(userID, bookID).Scan(&book.ID, &book.Title, &book.EReaderID, &book.UserID)
	if err != nil {
		WriteLog(LogError, err.Error())
	}
	annotations, err := GetAnnotations(userID, book.ID)
	if err != nil {
		book.Annotations = []Annotation{}
	} else {
		book.Annotations = annotations
	}
	return book
}

func CreateBook(eReaderID string, userID int, title string) (int, error) {
	db := database.GetDB()
	statement, err := db.Prepare("INSERT INTO books (title, ereader_id, user_id) VALUES (?, ?, ?)")
	if err != nil {
		return -1, err
	}
	res, err := statement.Exec(title, eReaderID, userID)
	if err != nil {
		return -1, err
	}
	bookID, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}
	return int(bookID), err
}

func BookExists(eReaderID string, userID int) int {
	db := database.GetDB()
	stmt, err := db.Prepare("SELECT id FROM books WHERE user_id = ? AND ereader_id = ?")
	if err != nil {
		WriteLog(LogError, err.Error())
		return -1
	}
	var bookID int
	err = stmt.QueryRow(userID, eReaderID).Scan(&bookID)
	if err != nil {
		WriteLog(LogError, err.Error())
		return -1
	}
	return bookID
}
