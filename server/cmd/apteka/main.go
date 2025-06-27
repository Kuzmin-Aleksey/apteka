package main

import (
	"log"
	"server/internal/app"
	"server/internal/config"
)

const configPath = "config/config.yaml"

func main() {
	cfg, err := config.ReadConfig(configPath)
	if err != nil {
		log.Fatal("read config file error:", err)
	}

	app.Run(cfg)
}
