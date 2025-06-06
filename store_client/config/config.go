package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	Service *ServiceConfig `yaml:"service"`
	Log     *LogConfig     `yaml:"log"`
}

type ServiceConfig struct {
	ServerAddr string `yaml:"server_addr"`
	ApiKey     string `yaml:"api_key"`
	StoreId    int    `yaml:"store_id"`
}

type LogConfig struct {
	EnableFileLog bool   `yaml:"enable_file_log"`
	OutputPath    string `yaml:"output_path"`
	LifeTimeDays  int    `yaml:"life_time_days"`
}

func ReadConfig(path string) (*Config, error) {
	config := new(Config)
	if err := cleanenv.ReadConfig(path, config); err != nil {
		return nil, err
	}
	return config, nil
}
