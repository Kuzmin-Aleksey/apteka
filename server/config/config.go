package config

import (
	"errors"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"os"
	"time"
)

type Config struct {
	Http      *HttpConfig      `json:"http" yaml:"http"`
	Api       *ApiConfig       `json:"api" yaml:"api"`
	Auth      *AuthConfig      `json:"auth" yaml:"auth"`
	DB        *DBConfig        `json:"db" yaml:"db"`
	Sphinx    *SphinxConfig    `json:"sphinx" yaml:"sphinx"`
	Redis     *RedisConfig     `json:"redis" yaml:"redis"`
	Web       *WebConfig       `json:"web" yaml:"web"`
	Images    *ImagesConfig    `json:"images" yaml:"images"`
	Promotion *PromotionConfig `json:"promotion" yaml:"promotion"`
}

type HttpConfig struct {
	Address         string `json:"address" yaml:"address"`
	ReadTimeoutSec  int    `json:"read_timeout_sec" yaml:"read_timeout_sec"`
	WriteTimeoutSec int    `json:"write_timeout_sec" yaml:"write_timeout_sec"`
	SSLKeyPath      string `json:"ssl_key_path" yaml:"ssl_key_path"`
	SSLCertPath     string `json:"ssl_cert_path" yaml:"ssl_cert_path"`
}

type ApiConfig struct {
	ApiKey           string `json:"api_key" yaml:"api_key"`
	HandleTimeoutSec int    `json:"handle_timeout_sec" yaml:"handle_timeout_sec"`
	CacheTemplate    bool   `json:"cache_template" yaml:"cache_template"`
}

type AuthConfig struct {
	Username     string `json:"username" yaml:"username"`
	Password     string `json:"password" yaml:"password"`
	TokenTTLDays int    `json:"token_ttl_days" yaml:"token_ttl_days"`
}

type DBConfig struct {
	Addr              string `json:"addr" yaml:"addr"`
	User              string `json:"user" yaml:"user"`
	Password          string `json:"password" yaml:"password"`
	Schema            string `json:"schema" yaml:"schema"`
	ConnectTimeoutSec int    `json:"connect_timeout_sec" yaml:"connect_timeout_sec"`
}

type SphinxConfig struct {
	Addr  string `json:"addr" yaml:"addr"`
	Index string `json:"index" yaml:"index"`
}

type RedisConfig struct {
	Host     string `json:"host" yaml:"host"`
	Password string `json:"password" yaml:"password"`
	DB       int    `json:"db" yaml:"db"`
}

type WebConfig struct {
	Title   string `json:"title" yaml:"title"`
	Logo    string `json:"logo" yaml:"logo"`
	LogoMin string `json:"logo_min" yaml:"logo_min"`
}

type PromotionConfig struct {
	AutoDelete bool `json:"auto-delete" yaml:"auto-delete"`
}

type ImagesConfig struct {
	FileRoot      string        `json:"file_root" yaml:"file_root"`
	AutoLoadDelay time.Duration `json:"auto_load_delay" yaml:"auto_load_delay"`
}

func ReadConfig(path string, dotenv ...string) (*Config, error) {
	if err := godotenv.Load(dotenv...); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	cfg := new(Config)
	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
