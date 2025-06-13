package products

import (
	"context"
	"fmt"
	"server/internal/domain/aggregate"
	"server/internal/domain/entity"
	"server/pkg/failure"
	"server/pkg/tx_manager"
	"slices"
	"time"
)

type ProductRepo interface {
	GetAll(ctx context.Context) ([]entity.Product, error)
	FindByCode(ctx context.Context, storeId int, id int) (*entity.Product, error)
	FindByCodes(ctx context.Context, storeId int, ids []int) ([]entity.Product, error)
	DeleteByStoreId(ctx context.Context, storeID int) error
	Save(ctx context.Context, product *entity.Product) error
}

type Searcher interface {
	Search(ctx context.Context, storeId int, searchString string) ([]aggregate.SearchResult, error)
}

type StoreRepo interface {
	SetLastUploadTime(ctx context.Context, storeId int, time time.Time) error
}

type BookingRepo interface {
	GetActiveProductsByStore(ctx context.Context, storeId int) ([]entity.BookProduct, error)
}

type ProductService struct {
	productRepo ProductRepo
	searcher    Searcher
	storeRepo   StoreRepo
	bookingRepo BookingRepo
}

func NewProductService(
	productRepo ProductRepo,
	searcher Searcher,
	storeRepo StoreRepo,
	bookingRepo BookingRepo,
) *ProductService {
	return &ProductService{
		productRepo: productRepo,
		searcher:    searcher,
		storeRepo:   storeRepo,
		bookingRepo: bookingRepo,
	}
}

func (s *ProductService) GetAll(ctx context.Context) ([]entity.Product, error) {
	return s.productRepo.GetAll(ctx)
}

func (s *ProductService) Search(ctx context.Context, storeId int, searchString string) ([]entity.Product, error) {
	results, err := s.searcher.Search(ctx, storeId, searchString)
	if err != nil {
		return nil, err
	}

	slices.SortFunc(results, func(a, b aggregate.SearchResult) int {
		return b.Compare(a)
	})

	products := make([]entity.Product, len(results))

	for i, result := range results {
		product, err := s.productRepo.FindByCode(ctx, storeId, result.Code)
		if err != nil {
			if failure.IsNotFoundError(err) {
				continue
			}
			return nil, err
		}

		products[i] = *product
	}

	products, err = s.filterProductsByBooking(ctx, storeId, products)
	if err != nil {
		return nil, err
	}

	return products, nil
}

func (s *ProductService) filterProductsByBooking(ctx context.Context, storeId int, products []entity.Product) ([]entity.Product, error) {
	filtered := make([]entity.Product, 0, len(products))

	bookProducts, err := s.bookingRepo.GetActiveProductsByStore(ctx, storeId)
	if err != nil {
		return nil, err
	}

	for _, product := range products {

		for _, bookProduct := range bookProducts {
			if bookProduct.CodeSTU == product.CodeSTU {
				product.Count -= bookProduct.Quantity
			}
		}

		if product.Count > 0 {
			filtered = append(filtered, product)
		}

	}

	return filtered, nil
}

func (s *ProductService) FindByIdS(ctx context.Context, storeId int, ids []int) (map[int]entity.Product, error) {
	products, err := s.productRepo.FindByCodes(ctx, storeId, ids)
	if err != nil {
		return nil, err
	}

	products, err = s.filterProductsByBooking(ctx, storeId, products)
	if err != nil {
		return nil, err
	}

	mapProducts := make(map[int]entity.Product)
	for _, product := range products {
		mapProducts[product.CodeSTU] = product
	}

	return mapProducts, nil
}

func (s *ProductService) CheckInStock(ctx context.Context, storeId int, checkingProducts []CheckInStockProduct) ([]CheckInStockProduct, error) {
	result := make([]CheckInStockProduct, 0, len(checkingProducts))

	bookProducts, err := s.bookingRepo.GetActiveProductsByStore(ctx, storeId)
	if err != nil {
		return nil, err
	}

	ids := make([]int, len(checkingProducts))

	for i, p := range checkingProducts {
		ids[i] = p.CodeSTU
	}

	products, err := s.productRepo.FindByCodes(ctx, storeId, ids)
	productsMap := make(map[int]entity.Product)
	for _, product := range products {
		productsMap[product.CodeSTU] = product
	}

	for _, checkingProduct := range checkingProducts {
		product, ok := productsMap[checkingProduct.CodeSTU]
		if !ok {
			continue
		}

		for _, bookProduct := range bookProducts {
			if product.CodeSTU == bookProduct.CodeSTU {
				product.Count -= bookProduct.Quantity
			}
		}

		checkingProduct.Count = product.Count

		if checkingProduct.Count > 0 {
			result = append(result, checkingProduct)
		}
	}

	return result, nil
}

func (s *ProductService) UploadProducts(ctx context.Context, products []entity.Product) error {
	if len(products) == 0 {
		return failure.NewInvalidRequestError("products is empty")
	}
	ctx, err := tx_manager.WithTx(ctx, s.productRepo)
	if err != nil {
		return failure.NewInternalError(err.Error())
	}

	if err := s.productRepo.DeleteByStoreId(ctx, products[0].StoreId); err != nil {
		if rbErr := tx_manager.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("%w, rollback err: %s", err, rbErr)
		}
		return err
	}

	for _, product := range products {
		if product.Price < 100 {
			continue
		}

		if err = s.productRepo.Save(ctx, &product); err != nil {
			if rbErr := tx_manager.Rollback(ctx); rbErr != nil {
				return fmt.Errorf("%w, rollback err: %s", err, rbErr)
			}
			return err
		}
	}

	if err := s.storeRepo.SetLastUploadTime(ctx, products[0].StoreId, time.Now()); err != nil {
		return err
	}

	if err := tx_manager.Commit(ctx); err != nil {
		return failure.NewInternalError(err.Error())
	}

	return nil
}
