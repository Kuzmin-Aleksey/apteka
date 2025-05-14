package repo

import (
	"apteka_client/models"
	"database/sql"
	"fmt"
	"os"
)

type ProductsRepo struct {
	scripts map[string]*sql.Stmt
	db      *sql.DB
}

func NewProductsRepo(db *sql.DB) (*ProductsRepo, error) {
	scripts := make(map[string]*sql.Stmt)
	var err error
	scripts["get_products"], err = readScript(db, "scripts/get_products.sql")
	if err != nil {
		return nil, fmt.Errorf("read script error: %s", err)
	}

	return &ProductsRepo{
		db:      db,
		scripts: scripts,
	}, nil
}

func (r *ProductsRepo) GetProducts(storeId int) ([]models.Product, error) {
	var products []models.Product

	rows, err := r.scripts["get_products"].Query(storeId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.CodeSTU, &p.GTIN, &p.Name, &p.Count, &p.Price, &p.Producer, &p.Country, &p.Description); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func readScript(db *sql.DB, filename string) (*sql.Stmt, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return db.Prepare(string(file))
}
