package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DBUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}
	configJSONString, err := os.ReadFile(configFilePath)
	if err != nil {
		return Config{}, err
	}
	var config Config
	err = json.Unmarshal(configJSONString, &config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}

func (c *Config) SetUser(username string) error {
	c.CurrentUserName = username
	err := write(*c)
	if err != nil {
		return err
	}
	return nil
}

func write(cfg Config) error {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return err
	}
	bytes, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(configFilePath, bytes, 0644)
}

func getConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configFileName), nil
}
