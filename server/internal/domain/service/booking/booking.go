package booking

import (
	"fmt"
	"golang.org/x/net/context"
	"io"
	"math"
	"os"
	"server/internal/domain/entity"
	"server/pkg/failure"
	"time"
)

type BookingRepo interface {
	Save(ctx context.Context, book *entity.Book) error
	UpdateStatus(ctx context.Context, bookId int, status string) error
	GetById(ctx context.Context, bookId int) (*entity.Book, error)
	GetByStore(ctx context.Context, storeId int) ([]entity.Book, error)
	Delete(ctx context.Context, bookId int) error
}

type StoreProvider interface {
	IsBookingEnable(ctx context.Context, storeId int) (bool, error)
}

type ProductsProvider interface {
	FindByIdS(ctx context.Context, storeId int, ids []int) (map[int]entity.Product, error)
}

type PromotionsProvider interface {
	GetAll(ctx context.Context) ([]entity.Promotion, error)
}

type BookingService struct {
	repo         BookingRepo
	store        StoreProvider
	products     ProductsProvider
	promotions   PromotionsProvider
	bookingDelay time.Duration
}

const bookingDelayFilePath = "config/booking-delay.txt"
const defaultBookingDelay = time.Hour * 4

func NewBookingService(repo BookingRepo, store StoreProvider, products ProductsProvider, promotions PromotionsProvider) (*BookingService, error) {
	s := &BookingService{
		repo:       repo,
		products:   products,
		promotions: promotions,
		store:      store,
	}

	f, err := os.OpenFile(bookingDelayFilePath, os.O_RDONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, err
	}

	delayBytes, err := io.ReadAll(f)
	if err != nil || len(delayBytes) == 0 {
		if err := s.StoreBookingDelay(defaultBookingDelay); err != nil {
			return nil, err
		}
	} else {
		bookingDelay, err := time.ParseDuration(string(delayBytes))
		if err != nil {
			return nil, fmt.Errorf("booking delay format error: %s\ncheck file %s", err, bookingDelayFilePath)
		}
		s.bookingDelay = bookingDelay
	}

	return s, nil
}

func (s *BookingService) ToBook(ctx context.Context, storeId int, dto *CreateBookDTO) (int, error) {
	if len(dto.Products) == 0 {
		return 0, failure.NewInvalidRequestError("no products provided")
	}

	// check enable
	bookingEnabled, err := s.store.IsBookingEnable(ctx, storeId)
	if err != nil {
		return 0, err
	}
	if !bookingEnabled {
		return 0, failure.NewLockedError("booking is disabled")
	}
	// check in stock, set price
	ids := make([]int, 0, len(dto.Products))

	for _, p := range dto.Products {
		ids = append(ids, p.CodeSTU)
	}

	products, err := s.products.FindByIdS(ctx, storeId, ids)
	if err != nil {
		return 0, err
	}

	resultProducts := make([]entity.BookProduct, 0, len(dto.Products))

	for _, p := range dto.Products {
		existProduct, ok := products[p.CodeSTU]
		if !ok {
			continue
		}
		if p.Quantity > existProduct.Count {
			p.Quantity = existProduct.Count
		}
		p.Price = existProduct.Price
		resultProducts = append(resultProducts, p)
	}

	if len(resultProducts) == 0 {
		return 0, failure.NewInvalidRequestError("all products not in stock")
	}

	dto.Products = resultProducts

	// check promotions
	promotions, err := s.promotions.GetAll(ctx)
	if err != nil {
		return 0, err
	}

	for i, p := range dto.Products {
		for _, promotion := range promotions {
			if promotion.ProductCode == p.CodeSTU {
				dto.Products[i].Price = p.Price - (promotion.Discount * 100)
			}
		}
	}

	book := &entity.Book{
		StoreId:   storeId,
		CreatedAt: time.Now(),
		Status:    entity.BookStatusCreated,
		Username:  dto.Username,
		Phone:     dto.Phone,
		Message:   dto.Message,
		Products:  dto.Products,
	}

	if err := s.repo.Save(ctx, book); err != nil {
		return 0, err
	}

	return book.Id, nil
}

func (s *BookingService) UpdateStatus(ctx context.Context, bookId int, status string) error {
	if err := s.repo.UpdateStatus(ctx, bookId, status); err != nil {
		return err
	}

	return nil
}

func (s *BookingService) Get(ctx context.Context, bookId int) (*GetBookingResponseDTO, error) {
	book, err := s.repo.GetById(ctx, bookId)
	if err != nil {
		return nil, err
	}

	return &GetBookingResponseDTO{
		Book:  *book,
		Delay: int(math.Round(s.bookingDelay.Hours())),
	}, nil
}

func (s *BookingService) GetByStore(ctx context.Context, storeId int) ([]entity.Book, error) {
	bookings, err := s.repo.GetByStore(ctx, storeId)
	if err != nil {
		return nil, err
	}
	return bookings, nil
}

func (s *BookingService) Delete(ctx context.Context, bookId int) error {
	if err := s.repo.Delete(ctx, bookId); err != nil {
		return err
	}
	return nil
}

func (s *BookingService) StoreBookingDelay(newDelay time.Duration) error {
	if newDelay == 0 {
		return failure.NewInvalidRequestError("booking delay is zero")
	} else if newDelay == s.bookingDelay {
		return nil
	}

	if err := os.WriteFile(bookingDelayFilePath, []byte(newDelay.String()), os.ModePerm); err != nil {
		return failure.NewInternalError(err.Error())
	}

	s.bookingDelay = newDelay

	return nil
}

func (s *BookingService) GetBookingDelay() time.Duration {
	return s.bookingDelay
}
