package app

import (
	"bytes"
	"efarma_integration/models"
	"efarma_integration/products_encoder"
	"efarma_integration/utils"
	"io"
	"log"
)

type ProductRepo interface {
	GetProducts(storeId int) ([]models.Product, error)
}

type Client interface {
	Unload(data io.Reader) error
}

type App struct {
	repo    ProductRepo
	client  Client
	storeId int
}

func NewApp(repo ProductRepo, client Client, storeId int) *App {
	return &App{
		repo:    repo,
		client:  client,
		storeId: storeId,
	}
}

func (app *App) Run() {
	products, err := app.repo.GetProducts(app.storeId)
	if err != nil {
		log.Fatal(err)
	}

	products = utils.MergingDuplicates(products)

	log.Printf("%d products found\n", len(products))

	buf := new(bytes.Buffer)

	if err := products_encoder.NewEncoder(buf).Encode(app.storeId, products); err != nil {
		log.Fatal(err)
	}

	if err := app.client.Unload(buf); err != nil {
		log.Fatal(err)
	}
}
