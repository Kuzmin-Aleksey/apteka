package promotion

import (
	"fmt"
	"golang.org/x/net/context"
	"io"
	"log"
	"server/internal/domain/entity"
	"server/internal/domain/service/promotion/promotion_parser"
	"server/pkg/failure"
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
	FindByCode(ctx context.Context, storeId int, code int) (*entity.Product, error)
}

type Logger interface {
	Println(v ...any)
	Printf(format string, v ...any)
}

type PromotionService struct {
	repo              PromotionRepo
	productsRepo      ProductsRepo
	l                 Logger
	shutdown          chan struct{}
	autoDeleteRunning bool
}

func NewPromotionService(repo PromotionRepo, products ProductsRepo, l Logger) *PromotionService {
	if l == nil {
		l = log.Default()
	}
	m := &PromotionService{
		repo:         repo,
		productsRepo: products,
		l:            l,
		shutdown:     make(chan struct{}),
	}
	return m
}

func (s *PromotionService) RunAutoDeletion() {
	s.autoDeleteRunning = true

	defer func() {
		s.autoDeleteRunning = false
	}()

	for {
		now := time.Now()
		nextUpdate := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)
		ticker := time.Tick(nextUpdate.Sub(now))

		select {
		case <-s.shutdown:
			return
		case <-ticker:
		}

		if err := s.repo.DeleteAll(context.Background()); err != nil {
			s.l.Println("failed to delete all promotions: ", err)
		}
	}
}

func (s *PromotionService) Shutdown(ctx context.Context) error {
	if s.autoDeleteRunning {
		select {
		case s.shutdown <- struct{}{}:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (s *PromotionService) UploadPromotionDocument(ctx context.Context, doc io.Reader) ([]entity.Promotion, error) {
	promotions, err := promotion_parser.ParseDoc(doc)
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

	for _, promotion := range promotions {
		if err := s.repo.Save(ctx, &promotion); err != nil {
			if rbErr := tx_manager.Rollback(ctx); rbErr != nil {
				return nil, fmt.Errorf("%w, rollback err: %s", err, rbErr)
			}
			return nil, err
		}
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

	for _, promotion := range allPromo {
		select {
		case <-ctx.Done():
			return nil, failure.NewInternalError("context err: " + ctx.Err().Error())
		default:
		}

		promotionInStock := PromotionInStock{
			Promotion: promotion,
			InStock:   true,
		}

		prod, err := s.productsRepo.FindByCode(ctx, storeId, promotion.ProductCode)
		if err != nil {
			if failure.IsNotFoundError(err) {
				promotionInStock.InStock = false
			}

			s.l.Printf("Promotion service: failed to find product <%d>: %s", promotion.ProductCode, err)
		} else {
			promotionInStock.PriceWithoutDiscount = prod.Price
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
