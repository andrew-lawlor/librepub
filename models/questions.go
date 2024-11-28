package models

import "github.com/andrew-lawlor/librepub/database"

type Question struct {
	ID       int
	Question string
	Answer   string
}

func NewQuestion(id int, question, answer string) Question {
	return Question{
		ID:       id,
		Question: question,
		Answer:   answer,
	}
}

func GetQuestion(id int) (Question, error) {
	db := database.GetDB()
	var query = "SELECT * FROM questions WHERE id = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		WriteLog(LogError, err.Error())
	}
	row := stmt.QueryRow(id)
	var question Question
	err = row.Scan(&question.ID, &question.Question, &question.Answer)
	if err != nil {
		WriteLog(LogError, err.Error())
	}
	return question, err
}

func GetRandomQuestion() (Question, error) {
	db := database.GetDB()
	var query = "SELECT * FROM questions ORDER BY RANDOM() LIMIT 1"
	stmt, err := db.Prepare(query)
	if err != nil {
		WriteLog(LogError, err.Error())
	}
	row := stmt.QueryRow()
	var question Question
	err = row.Scan(&question.ID, &question.Question, &question.Answer)
	if err != nil {
		WriteLog(LogError, err.Error())
	}
	return question, err
}
