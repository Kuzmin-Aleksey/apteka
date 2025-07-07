package main

import (
	"efarma_integration/DB"
	"efarma_integration/app"
	"efarma_integration/config"
	"efarma_integration/repo"
	"efarma_integration/unloadAdapter/client"
	"efarma_integration/unloadAdapter/file"
	"io"
	"log"
	"os"
)

const ConfigPath = "config/config.json"
const LogPath = "logs/last_unload.log"

func main() {
	initLogger()

	cnf, err := config.GetConfig(ConfigPath)
	if err != nil {
		log.Fatal(err)
	}

	db, err := DB.Connect(&cnf.DB)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	productRepo, err := repo.NewProductsRepo(db)
	if err != nil {
		log.Fatal(err)
	}

	var c app.Client
	if len(os.Args) > 1 {
		log.Println("unload in file: ", os.Args[1])
		c = file.NewUnloadFile(os.Args[1])
	} else {
		c = client.New(&cnf.HttpClient)
	}

	a := app.NewApp(productRepo, c, cnf.DbStoreId, cnf.StoreId)

	a.Run()
}

func initLogger() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	f, err := os.OpenFile(LogPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(io.MultiWriter(f, os.Stdout))
}
