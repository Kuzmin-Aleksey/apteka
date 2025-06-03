package mysql

import (
	"database/sql"
	"errors"
	"golang.org/x/net/context"
	"server/internal/domain/entity"
	"server/pkg/failure"
)

type PromotionRepo struct {
	DB
}

func NewPromotion(db DB) *PromotionRepo {
	return &PromotionRepo{
		DB: db,
	}
}

func (r *PromotionRepo) Save(ctx context.Context, promotion *entity.Promotion) error {
	if _, err := r.DB.ExecContext(ctx, "INSERT INTO promotions (product_code, product_name, discount) VALUES (?, ?, ?)",
		promotion.ProductCode, promotion.ProductName, promotion.Discount); err != nil {
		return failure.NewInternalError(err.Error())
	}
	return nil
}

func (r *PromotionRepo) Update(ctx context.Context, promotion *entity.Promotion) error {
	if _, err := r.DB.ExecContext(ctx, "UPDATE promotions SET product_name=?, discount=? WHERE product_code=?",
		promotion.ProductName, promotion.Discount, promotion.ProductCode); err != nil {
		return failure.NewInternalError(err.Error())
	}
	return nil
}

func (r *PromotionRepo) DeleteAll(ctx context.Context) error {
	if _, err := r.DB.ExecContext(ctx, "DELETE FROM promotions"); err != nil {
		return failure.NewInternalError(err.Error())
	}
	return nil
}

func (r *PromotionRepo) Delete(ctx context.Context, productCode int) error {
	if _, err := r.DB.ExecContext(ctx, "DELETE FROM promotions WHERE product_code=?", productCode); err != nil {
		return failure.NewInternalError(err.Error())
	}
	return nil
}

func (r *PromotionRepo) GetAll(ctx context.Context) ([]entity.Promotion, error) {
	promotions := make([]entity.Promotion, 0, 1)

	rows, err := r.DB.QueryContext(ctx, "SELECT * FROM promotions")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return promotions, nil
		}
		return nil, failure.NewInternalError(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var promotion entity.Promotion
		if err := rows.Scan(&promotion.ProductCode, &promotion.ProductName, &promotion.Discount); err != nil {
			return nil, failure.NewInternalError(err.Error())
		}
		promotions = append(promotions, promotion)
	}
	return promotions, nil
}
