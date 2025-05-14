package main

import (
	"apteka_booking/config"
	"apteka_booking/service"
	"apteka_booking/ui"
	"log"
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
