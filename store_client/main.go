package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"store_client/config"
	"store_client/service"
	"store_client/ui"
	"time"
)

func main() {
	cfg, err := config.ReadConfig("config/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	if err := setupLogger(cfg.Log); err != nil {
		log.Println("setup logger error:", err)
	}

	s := service.NewService(cfg.Service)
	app := ui.New(s)
	app.Run()
}

func setupLogger(cfg *config.LogConfig) error {
	if cfg == nil {
		return nil
	}
	
	if cfg.EnableFileLog {
		if cfg.OutputPath == "" {
			cfg.OutputPath = "logs"
		}

		if err := os.MkdirAll(cfg.OutputPath, os.ModePerm); err != nil {
			return err
		}

		filePath := filepath.Join(cfg.OutputPath, time.Now().Format(time.DateOnly)+".log")
		logFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return err
		}

		log.SetOutput(io.MultiWriter(os.Stderr, logFile))

		// remove old files

		logFiles, err := os.ReadDir(cfg.OutputPath)
		if err != nil {
			return err
		}

		deleteBefore := time.Now().Add(time.Duration(-cfg.LifeTimeDays) * time.Hour * 24)

		for _, file := range logFiles {
			if file.IsDir() {
				continue
			}

			t, err := time.Parse(time.DateOnly+".log", file.Name())
			if err != nil {
				log.Println("parse file time error: ", err)
				continue
			}

			if t.Before(deleteBefore) {
				if err := os.Remove(file.Name()); err != nil {
					log.Println("remove file error: ", err)
				}
			}
		}

	}
	return nil
}
