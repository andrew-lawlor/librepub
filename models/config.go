package models

import "github.com/andrew-lawlor/librepub/database"

type Config struct {
	ID    int
	Name  string
	Value string
	Type  string
}

func NewConfig(id int, name string, value string, configType string) Config {
	return Config{
		ID:    id,
		Name:  name,
		Value: value,
		Type:  configType,
	}
}

func EditConfig(newConfig Config) bool {
	db := database.GetDB()
	statement, err := db.Prepare("UPDATE config SET value = ? WHERE id = ?")
	if err != nil {
		return false
	}
	_, err = statement.Exec(newConfig.Value, newConfig.ID)
	return err == nil
}

func GetConfig(id int) (Config, error) {
	db := database.GetDB()
	var query = "SELECT * FROM config WHERE id = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		WriteLog(LogError, err.Error())
	}
	row := stmt.QueryRow(id)
	var config Config
	err = row.Scan(&config.ID, &config.Name, &config.Value, &config.Type)
	if err != nil {
		WriteLog(LogError, err.Error())
	}
	return config, err
}

func GetConfigs() []Config {
	db := database.GetDB()
	stmt, err := db.Prepare("SELECT * FROM config")
	if err != nil {
		WriteLog(LogError, err.Error())
	}
	row, err := stmt.Query()
	if err != nil {
		WriteLog(LogError, err.Error())
	}
	var configs = []Config{}
	defer row.Close()
	for row.Next() { // Iterate and fetch the records from result cursor
		var config Config
		err = row.Scan(&config.ID, &config.Name, &config.Value, &config.Type)
		if err != nil {
			WriteLog(LogError, err.Error())
		}
		configs = append(configs, config)
	}
	return configs
}
