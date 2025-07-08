package promotion

import (
	"fmt"
	"golang.org/x/net/context"
	"io"
	"server/internal/domain/entity"
	"server/internal/domain/service/promotion/promotion_parser"
	"server/pkg/contextx"
	"server/pkg/failure"
	"server/pkg/logx"
	"server/pkg/tx_manager"
	"time"
)

type PromotionRepo interface {
	Save(ctx context.Context, promotions *entity.Promotion) error
	Update(ctx context.Context, promotion *entity.Promotion) error
	DeleteAll(ctx context.Context) error
	Delete(ctx context.Context, productCode int) error
	GetAll(ctx context.Context) ([]entity.Promotion, error)
}

type ProductsRepo interface {
	FindByIdS(ctx context.Context, storeId int, ids []int) (map[int]entity.Product, error)
}

type PromotionService struct {
	repo         PromotionRepo
	productsRepo ProductsRepo
}

func NewPromotionService(repo PromotionRepo, products ProductsRepo) *PromotionService {
	m := &PromotionService{
		repo:         repo,
		productsRepo: products,
	}
	return m
}

func (s *PromotionService) RunAutoDeletion(ctx context.Context) {
	l := contextx.GetLoggerOrDefault(ctx)

	for {
		now := time.Now()
		nextUpdate := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)
		ticker := time.Tick(nextUpdate.Sub(now))

		select {
		case <-ctx.Done():
			l.Info("auto delete promotion stop", logx.Error(ctx.Err()))
			return
		case <-ticker:
		}

		if err := s.repo.DeleteAll(context.Background()); err != nil {
			l.Error("delete all products failed", logx.Error(err))
		}
	}
}

func (s *PromotionService) UploadPromotionDocument(ctx context.Context, doc io.Reader) ([]entity.Promotion, error) {
	promotions, err := promotion_parser.ParseDoc(ctx, doc)
	if err != nil {
		return nil, err
	}

	ctx, err = tx_manager.WithTx(ctx, s.repo)
	if err != nil {
		return nil, failure.NewInternalError(err.Error())
	}
	if err := s.repo.DeleteAll(ctx); err != nil {
		if rbErr := tx_manager.Rollback(ctx); rbErr != nil {
			return nil, fmt.Errorf("%w, rollback err: %s", err, rbErr)
		}
		return nil, err
	}

	existPromotions := make(map[int]struct{})

	for _, promotion := range promotions {
		if _, ok := existPromotions[promotion.ProductCode]; ok {
			contextx.GetLoggerOrDefault(ctx).Warn("UploadPromotionDocument - duplicate promotion", "promotion", promotion)
			continue
		}
		if err := s.repo.Save(ctx, &promotion); err != nil {
			if rbErr := tx_manager.Rollback(ctx); rbErr != nil {
				return nil, fmt.Errorf("%w, rollback err: %s", err, rbErr)
			}
			return nil, err
		}
		existPromotions[promotion.ProductCode] = struct{}{}
	}

	if err := tx_manager.Commit(ctx); err != nil {
		return nil, failure.NewInternalError(err.Error())
	}

	return promotions, nil
}

func (s *PromotionService) New(ctx context.Context, promotion *entity.Promotion) error {
	if err := s.repo.Save(ctx, promotion); err != nil {
		return err
	}
	return nil
}

func (s *PromotionService) Update(ctx context.Context, promotion *entity.Promotion) error {
	if err := s.repo.Update(ctx, promotion); err != nil {
		return err
	}
	return nil
}

func (s *PromotionService) GetAll(ctx context.Context) ([]entity.Promotion, error) {
	p, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return p, err
}

func (s *PromotionService) GetInStock(ctx context.Context, storeId int) ([]PromotionInStock, error) {
	allPromo, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	inStock := make([]PromotionInStock, 0, len(allPromo)/2)

	ids := make([]int, 0, len(allPromo))

	for _, promotion := range allPromo {
		ids = append(ids, promotion.ProductCode)
	}

	products, err := s.productsRepo.FindByIdS(ctx, storeId, ids)
	if err != nil {
		return nil, err
	}

	for _, promotion := range allPromo {
		product, ok := products[promotion.ProductCode]

		promotionInStock := PromotionInStock{
			Promotion: promotion,
			InStock:   ok,
		}
		if ok {
			promotionInStock.Product = &product
		}

		inStock = append(inStock, promotionInStock)
	}

	return inStock, nil
}

func (s *PromotionService) DeleteAll(ctx context.Context) error {
	if err := s.repo.DeleteAll(ctx); err != nil {
		return err
	}
	return nil
}

func (s *PromotionService) Delete(ctx context.Context, productCode int) error {
	if err := s.repo.Delete(ctx, productCode); err != nil {
		return err
	}
	return nil
}
