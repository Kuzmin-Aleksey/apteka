package main

import (
	"log"
	"server/config"
	"server/internal/app"
)

const configPath = "config/config.yaml"

func main() {
	cfg, err := config.ReadConfig(configPath)
	if err != nil {
		log.Fatal("read config file error:", err)
	}

	app.Run(cfg)
}
