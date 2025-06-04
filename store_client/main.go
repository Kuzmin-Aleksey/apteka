package main

import (
	"log"
	"store_client/config"
	"store_client/service"
	"store_client/ui"
)

func main() {
	cfg, err := config.ReadConfig("config/config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	s := service.NewService(cfg.Service)
	app := ui.New(s)
	app.Run()
}
