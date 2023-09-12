package discord

import (
	"encoding/json"
	"fmt"
	"os"
)

var (
	Conf Config
)

type Config struct {
	DebugLog bool   `json:"debug_log"`
	Prefix   string `json:"prefix"`
	Token    string `json:"token"`
}

func LoadConfig(filename string) (*Config, error) {
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("cannot loading config. %v", err)
	}

	if err := json.Unmarshal(body, &Conf); err != nil {
		return nil, fmt.Errorf("cannot unmarshal config. %v", err)
	}

	return &Conf, nil
}
