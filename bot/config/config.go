package config

import (
	"encoding/json"
	"os"
)

// BotConfiguration configures a JapanBot
type BotConfiguration struct {
	JMdictFile string `json:"jmdict_file"`
	APIToken   string `json:"api_token"`
}

// LoadFromFile creates a BotConfiguration from a given file
func LoadFromFile(file string) (*BotConfiguration, error) {
	s, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	bc := &BotConfiguration{}
	dec := json.NewDecoder(s)
	for dec.More() {
		if err = dec.Decode(bc); err != nil {
			return nil, err
		}
	}

	return bc, nil
}
