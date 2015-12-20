package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	DbConnStr string `json:"dbConnStr"`
}

func NewConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
