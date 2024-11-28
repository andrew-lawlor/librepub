package models

import (
	"database/sql"
	"html/template"
	"log"
	"strconv"

	"github.com/andrew-lawlor/librepub/database"
)

type Vocab struct {
	ID          int
	Word        string
	Definition  string
	BookTitle   string
	ContentHTML template.HTML
	Created     string
}

func GetVocab(page int, search string, userID int) []Vocab {
	var offset = (page - 1) * page_offset
	db := database.GetDB()
	var query string
	var queryArg string
	if search == "" {
		query = "SELECT id, word, definition, created FROM vocab WHERE user_id = ? ORDER BY created DESC LIMIT 5 OFFSET ?"
		queryArg = strconv.Itoa(offset)
	} else {
		query = "SELECT id, word, definition, created FROM vocab WHERE user_id = ? AND word LIKE ? ORDER BY created DESC LIMIT 5 OFFSET " + strconv.Itoa(offset)
		queryArg = "%" + search + "%"
	}
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}
	row, err := stmt.Query(userID, queryArg)
	if err != nil {
		log.Fatal(err)
	}
	var vocabList = []Vocab{}
	defer row.Close()
	for row.Next() { // Iterate and fetch the records from result cursor
		var vocab Vocab
		err = row.Scan(&vocab.ID, &vocab.Word, &vocab.Definition, &vocab.Created)
		if err != nil {
			WriteLog(LogError, err.Error())
			continue
		}
		vocab.ContentHTML = parseHTMLString(vocab.Definition)
		vocabList = append(vocabList, vocab)
	}
	return vocabList
}

func GetAllVocab(userID int) []Vocab {
	db := database.GetDB()
	stmt, err := db.Prepare("SELECT word, definition FROM vocab WHERE user_id = ? ORDER BY created")
	if err != nil {
		log.Fatal(err)
	}
	row, err := stmt.Query(userID)
	if err != nil {
		log.Fatal(err)
	}
	var vocabList = []Vocab{}
	defer row.Close()
	for row.Next() { // Iterate and fetch the records from result cursor
		var vocab Vocab
		err = row.Scan(&vocab.Word, &vocab.Definition)
		if err != nil {
			WriteLog(LogError, err.Error())
			continue
		}
		vocab.ContentHTML = parseHTMLString(vocab.Definition)
		vocabList = append(vocabList, vocab)
	}
	return vocabList
}

func DeleteVocab(id int, userID int) bool {
	db := database.GetDB()
	statement, _ := db.Prepare("DELETE from vocab WHERE id = ? AND user_id = ?")
	_, err := statement.Exec(id, userID)
	return err == nil
}

// Extracts vocab data from Kobo DB.
func ExportVocab(dbPath string) ([]Vocab, error) {
	var vocab = []Vocab{}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		WriteLog(LogError, err.Error())
		return vocab, err
	}
	stmt, err := db.Prepare("SELECT Text, VolumeID, DateCreated FROM WordList")
	if err != nil {
		WriteLog(LogError, err.Error())
		return vocab, err
	}
	row, err := stmt.Query()
	if err != nil {
		WriteLog(LogError, err.Error())
		return vocab, err
	}
	defer row.Close()
	for row.Next() {
		var vocabEntry Vocab
		var eReaderID string
		err = row.Scan(&vocabEntry.Word, &eReaderID, &vocabEntry.Created)
		if err != nil {
			WriteLog(LogError, err.Error())
			return vocab, err
		}
		vocabEntry.BookTitle = GetBookTitle(dbPath, eReaderID)
		vocabEntry.Definition = lookupWord(vocabEntry.Word)
		vocab = append(vocab, vocabEntry)
	}
	return vocab, nil
}

// Writes Kobo DB data to our DB.
func ImportVocab(vocab []Vocab, userID int) error {
	db := database.GetDB()
	statement, err := db.Prepare("INSERT INTO vocab (word, definition, created, user_id) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	for _, v := range vocab {
		if vocabExists(v.Word, userID) {
			continue
		}
		_, err := statement.Exec(v.Word, v.Definition, v.Created, userID)
		if err != nil {
			return err
		}
	}
	return nil
}

// Looks up word in a sqlite DB created from Wiktionary.
func lookupWord(word string) string {
	db, err := sql.Open("sqlite3", "./dicts/en/dict-en.db")
	if err != nil {
		WriteLog(LogError, err.Error())
		return ""
	}
	stmt, err := db.Prepare("SELECT definition FROM words WHERE word = ?")
	if err != nil {
		WriteLog(LogError, err.Error())
	}
	var definition string
	err = stmt.QueryRow(word).Scan(&definition)
	if err != nil {
		WriteLog(LogError, err.Error())
		return ""
	}
	return definition
}

func vocabExists(word string, userID int) bool {
	db := database.GetDB()
	stmt, err := db.Prepare("SELECT word FROM vocab WHERE word = ? AND user_id = ?")
	if err != nil {
		WriteLog(LogError, err.Error())
		return false
	}
	var id string
	err = stmt.QueryRow(word, userID).Scan(&id)
	if err != nil {
		WriteLog(LogError, err.Error())
		return false
	}
	return true
}

// Render definition as template, since it contains HTML.
func parseHTMLString(htmlString string) template.HTML {
	return template.HTML(htmlString)
}
