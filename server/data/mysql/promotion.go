package mysql

import (
	"database/sql"
	"errors"
	"golang.org/x/net/context"
	"server/domain/models"
)

type PromotionRepo struct {
	DB
}

func NewPromotion(db DB) *PromotionRepo {
	return &PromotionRepo{
		DB: db,
	}
}

func (r *PromotionRepo) Save(ctx context.Context, promotion *models.Promotion) error {
	if _, err := r.DB.ExecContext(ctx, "INSERT INTO promotions (product_code, product_name, discount) VALUES (?, ?, ?)",
		promotion.ProductCode, promotion.ProductName, promotion.Discount); err != nil {
		return models.NewError(models.ErrDatabaseError, err)
	}
	return nil
}

func (r *PromotionRepo) Update(ctx context.Context, promotion *models.Promotion) error {
	if _, err := r.DB.ExecContext(ctx, "UPDATE promotions SET product_name=?, discount=? WHERE product_code=?",
		promotion.ProductName, promotion.Discount, promotion.ProductCode); err != nil {
		return models.NewError(models.ErrDatabaseError, err)
	}
	return nil
}

func (r *PromotionRepo) DeleteAll(ctx context.Context) error {
	if _, err := r.DB.ExecContext(ctx, "DELETE FROM promotions"); err != nil {
		return models.NewError(models.ErrDatabaseError, err)
	}
	return nil
}

func (r *PromotionRepo) Delete(ctx context.Context, productCode int) error {
	if _, err := r.DB.ExecContext(ctx, "DELETE FROM promotions WHERE product_code=?", productCode); err != nil {
		return models.NewError(models.ErrDatabaseError, err)
	}
	return nil
}

func (r *PromotionRepo) GetAll(ctx context.Context) ([]models.Promotion, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT * FROM promotions")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, models.NewError(models.ErrDatabaseError, err)
	}
	defer rows.Close()

	promotions := make([]models.Promotion, 0, 1)
	for rows.Next() {
		var promotion models.Promotion
		if err := rows.Scan(&promotion.ProductCode, &promotion.ProductName, &promotion.Discount); err != nil {
			return nil, models.NewError(models.ErrDatabaseError, err)
		}
		promotions = append(promotions, promotion)
	}
	return promotions, nil
}
