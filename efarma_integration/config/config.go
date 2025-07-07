package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	DbStoreId  int              `json:"db_store_id"`
	StoreId    int              `json:"store_id"`
	HttpClient HttpClientConfig `json:"http_client"`
	DB         DbConfig         `json:"db"`
}

type HttpClientConfig struct {
	UnloadUrl string `json:"unload_url"`
	Timeout   int    `json:"timeout"`
	ApiKey    string `json:"api_key"`
}

type DbConfig struct {
	Host       string         `json:"host"`
	Username   string         `json:"username"`
	Password   string         `json:"password"`
	DBName     string         `json:"db_name"`
	Args       map[string]any `json:"args"`
	SQLExpress bool           `json:"sql_express"`
}

func GetConfig(configPath string) (*Config, error) {
	config := new(Config)

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(config); err != nil {
		return nil, fmt.Errorf("config file invalid: %v", err)
	}

	return config, nil
}
