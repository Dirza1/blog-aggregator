package config

import (
	"encoding/json"
	"os"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	Url  string `json:"db_url"`
	User string `json:"current_user_name"`
}

func Read() Config {
	config := Config{}
	filepath, err := getConfigFilePath()
	if err != nil {
		panic(err)
	}

	val, err := os.ReadFile(filepath + configFileName)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(val, &config)
	if err != nil {
		panic(err)
	}

	return config
}

func (c Config) SetUser(user string) {
	c.User = user
	write(c)
}

func getConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	return home, nil
}

func write(c Config) error {
	bytes, err := json.Marshal(c)
	if err != nil {
		return err
	}
	err = os.WriteFile(".gatorconfig.json", bytes, 06647)
	if err != nil {
		return err
	}
	return nil
}
