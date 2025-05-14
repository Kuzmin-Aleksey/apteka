package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	Service *ServiceConfig `yaml:"service"`
}

type ServiceConfig struct {
	ServerAddr string `yaml:"server_addr"`
	ApiKey     string `yaml:"api_key"`
	StoreId    int    `yaml:"store_id"`
}

func ReadConfig(path string) (*Config, error) {
	config := new(Config)
	if err := cleanenv.ReadConfig(path, config); err != nil {
		return nil, err
	}
	return config, nil
}
