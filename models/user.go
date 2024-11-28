package models

import (
	"database/sql"
	"time"

	"github.com/andrew-lawlor/librepub/database"
)

type User struct {
	UserID      int
	UserName    string
	DisplayName string
	Password    string
	Created     string
}

func NewUser(userName string, displayName string, hashedPassword string) bool {
	var user = User{
		UserName:    userName,
		DisplayName: displayName,
		Password:    hashedPassword,
		Created:     time.Now().Format(time.RFC3339),
	}
	return createUser(user)
}

func createUser(user User) bool {
	db := database.GetDB()
	// Create
	statement, err := db.Prepare("INSERT INTO users (user_name, password, display_name, created) VALUES (?, ?, ?, ?)")
	if err != nil {
		WriteLog(LogError, err.Error())
		return false
	}
	_, err = statement.Exec(user.UserName, user.Password, user.DisplayName, user.Created)
	if err != nil {
		WriteLog(LogError, err.Error())
		return false
	}
	WriteLog(LogInfo, "User registered: "+user.UserName)
	return true
}

func GetUser(userName string) (User, error) {
	db := database.GetDB()

	stmt, err := db.Prepare("SELECT * FROM users WHERE user_name = ?")
	if err != nil {
		WriteLog(LogError, err.Error())
	}
	defer stmt.Close()
	var user User
	err = stmt.QueryRow(userName).Scan(&user.UserName, &user.UserID, &user.Password, &user.DisplayName, &user.Created)
	if err != nil {
		if err == sql.ErrNoRows {
			// Handle the case of no rows returned.
			return user, err
		}
	}
	return user, err
}
