package Config

import (
	"encoding/json"
	"fmt"
	"os"
)

type (
	Config struct {
		Redis Redis   `json:"redis"`
		Token token   `json:"token"`
		Gpt   Gpt3    `json:"Gpt3"`
		Mongo MongoDb `json:"MongoDB"`
	}
	Redis struct {
		Addr     string `json:"addr"`
		Password string `json:"password"`
		DB       int    `json:"db"`
	}
	token struct {
		Token string `json:"token"`
	}
	Gpt3 struct {
		ApiKey string `json:"ApiKey"`
		Model  string `json:"model"`
	}
	MongoDb struct {
		Uri            string `json:"Uri"`
		DataBaseName   string `json:"DataBaseName"`
		CollectionName string `json:"collection_name"`
	}
)

func ParseConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %v", err)
	}
	var Cfg Config
	err = json.Unmarshal(file, &Cfg)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}
	return &Cfg, nil
}
