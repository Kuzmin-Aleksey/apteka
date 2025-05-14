package products

import (
	"context"
	"server/domain/models"
	"server/pkg/tx_manager"
	"time"
)

type ProductRepo interface {
	GetAll(ctx context.Context) ([]models.Product, error)
	Find(ctx context.Context, storeId int, searchString string) ([]models.Product, error)
	FindByCode(ctx context.Context, storeId int, id int) (*models.Product, error)
	DeleteByStoreId(ctx context.Context, storeID int) error
	Save(ctx context.Context, product *models.Product) error
}

type StoreRepo interface {
	SetLastUploadTime(ctx context.Context, storeId int, time time.Time) error
}

type BookingRepo interface {
	GetActiveProductsByStore(ctx context.Context, storeId int) ([]models.BookProduct, error)
}

type ProductService struct {
	productRepo ProductRepo
	storeRepo   StoreRepo
	bookingRepo BookingRepo
}

func NewProductService(
	productRepo ProductRepo,
	storeRepo StoreRepo,
	bookingRepo BookingRepo,
) *ProductService {
	return &ProductService{
		productRepo: productRepo,
		storeRepo:   storeRepo,
		bookingRepo: bookingRepo,
	}
}

func (s *ProductService) GetAll(ctx context.Context) ([]models.Product, error) {
	return s.productRepo.GetAll(ctx)
}

func (s *ProductService) Search(ctx context.Context, storeId int, searchString string) ([]models.Product, error) {
	products, err := s.productRepo.Find(ctx, storeId, searchString)
	if err != nil {
		return nil, err
	}

	products, err = s.filterProductsByBooking(ctx, storeId, products)
	if err != nil {
		return nil, err
	}

	return products, nil
}

func (s *ProductService) filterProductsByBooking(ctx context.Context, storeId int, products []models.Product) ([]models.Product, error) {
	filtered := make([]models.Product, 0, len(products))

	bookProducts, err := s.bookingRepo.GetActiveProductsByStore(ctx, storeId)
	if err != nil {
		return nil, err
	}

	for _, product := range products {

		for _, bookProduct := range bookProducts {
			if bookProduct.CodeSTU == product.CodeSTU {
				product.Count -= bookProduct.Quantity
				break
			}
		}

		if product.Count > 0 {
			filtered = append(filtered, product)
		}

	}

	return filtered, nil
}

func (s *ProductService) FindById(ctx context.Context, storeId int, id int) (*models.Product, error) {
	product, err := s.productRepo.FindByCode(ctx, storeId, id)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (s *ProductService) UploadProducts(ctx context.Context, products []models.Product) error {
	if len(products) == 0 {
		return models.NewError(models.ErrInvalidRequest, "products is empty")
	}
	ctx, err := tx_manager.WithTx(ctx, s.productRepo)
	if err != nil {
		return models.NewError(models.ErrDatabaseError, err)
	}

	if err := s.productRepo.DeleteByStoreId(ctx, products[0].StoreId); err != nil {
		return models.AddError(err, tx_manager.Rollback(ctx))
	}

	for _, product := range products {
		if product.Price < 100 {
			continue
		}

		if err = s.productRepo.Save(ctx, &product); err != nil {
			return models.AddError(err, tx_manager.Rollback(ctx))
		}
	}

	if err := tx_manager.Commit(ctx); err != nil {
		return models.NewError(models.ErrDatabaseError, err)
	}

	if err := s.storeRepo.SetLastUploadTime(ctx, products[0].StoreId, time.Now()); err != nil {
		return err
	}

	return nil
}
