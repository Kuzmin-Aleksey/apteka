package mysql

import (
	"database/sql"
	errorsutils "errors"
	"golang.org/x/net/context"
	"server/domain/models"
)

type ProductRepo struct {
	DB
}

func NewProductRepo(db DB) *ProductRepo {
	return &ProductRepo{
		DB: db,
	}
}

// GetAll StoreId always 0
func (r *ProductRepo) GetAll(ctx context.Context) ([]models.Product, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT Code, MAX(GTIN), MAX(Name), SUM(Count), MAX(Price), MAX(Producer), MAX(Country), MAX(Description) FROM products GROUP BY Code")
	if err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		if err := rows.Scan(
			&product.CodeSTU,
			&product.GTIN,
			&product.Name,
			&product.Count,
			&product.Price,
			&product.Producer,
			&product.Country,
			&product.Description); err != nil {
			return nil, models.NewError(models.ErrDatabaseError, err)
		}

		products = append(products, product)
	}

	return products, nil
}

func (r *ProductRepo) Find(ctx context.Context, storeId int, searchString string) ([]models.Product, error) {
	rows, err := r.DB.QueryContext(ctx, queryFindProducts, searchString, storeId, storeId)
	if err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return nil, models.NewError(models.ErrNotFound, "no result", searchString)
		}
		return nil, models.NewError(models.ErrDatabaseError, err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		p, err := r.scanProduct(rows)
		if err != nil {
			return nil, models.NewError(models.ErrDatabaseError, err)
		}
		products = append(products, p)
	}

	return products, nil
}

func (r *ProductRepo) Save(ctx context.Context, product *models.Product) error {
	if _, err := r.DB.ExecContext(ctx, "INSERT INTO products (Code, StoreID, GTIN, Name, Count, Price, Producer, Country, Description) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		product.CodeSTU, product.StoreId, product.GTIN, product.Name, product.Count, product.Price, product.Producer, product.Country, product.Description); err != nil {
		return models.NewError(models.ErrDatabaseError, err)
	}
	return nil
}

func (r *ProductRepo) FindByCode(ctx context.Context, storeId int, code int) (*models.Product, error) {
	var product models.Product
	row := r.DB.QueryRowContext(ctx, "SELECT * FROM products WHERE Code=? AND StoreID=? ", code, storeId)
	product, err := r.scanProduct(row)
	if err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return nil, models.NewError(models.ErrNotFound, "no result", code, storeId)
		}
		return nil, models.NewError(models.ErrDatabaseError, err)
	}
	return &product, nil
}

func (r *ProductRepo) DeleteByStoreId(ctx context.Context, storeId int) error {
	if _, err := r.DB.ExecContext(ctx, "DELETE FROM products WHERE StoreID=?", storeId); err != nil {
		return err
	}
	return nil
}

func (r *ProductRepo) scanProduct(s Scanner) (models.Product, error) {
	var product models.Product
	if err := s.Scan(
		&product.CodeSTU,
		&product.StoreId,
		&product.GTIN,
		&product.Name,
		&product.Count,
		&product.Price,
		&product.Producer,
		&product.Country,
		&product.Description); err != nil {
		return product, err
	}
	return product, nil
}

const queryFindProducts = `
WITH res AS (
SELECT
	Code,
	MATCH (Name,Description) AGAINST (?) AS score
FROM products
WHERE 
	StoreID = ?

ORDER BY score DESC
)

SELECT 
	p.Code, p.StoreID, p.GTIN, p.Name, p.Count, p.Price, p.Producer, p.Country, p.Description
FROM res r
inner join products p ON p.Code = r.Code AND p.StoreID = ?
WHERE r.score > 0;
`
