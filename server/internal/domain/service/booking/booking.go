package booking

import (
	"fmt"
	"golang.org/x/net/context"
	"io"
	"log/slog"
	"math"
	"os"
	"server/internal/domain/aggregate"
	"server/internal/domain/entity"
	"server/pkg/contextx"
	"server/pkg/failure"
	"server/pkg/logx"
	"time"
)

type BookingRepo interface {
	Save(ctx context.Context, book *aggregate.BookWithProducts) error
	UpdateStatus(ctx context.Context, bookId int, status string) error
	GetById(ctx context.Context, bookId int) (*aggregate.BookWithProducts, error)
	GetByIds(ctx context.Context, bookIds []int) ([]aggregate.BookWithProducts, error)
	GetByStore(ctx context.Context, storeId int) ([]aggregate.BookWithProducts, error)
	GetActive(ctx context.Context) ([]entity.Book, error)
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
			return nil, fmt.Errorf("booking delay format error: %w\ncheck file %s", err, bookingDelayFilePath)
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
				if promotion.IsPercent {
					price := dto.Products[i].Price
					dto.Products[i].Price = price - price*promotion.Discount/100
				} else {
					dto.Products[i].Price = p.Price - (promotion.Discount * 100)
				}
			}
		}
	}

	book := &aggregate.BookWithProducts{
		Book: entity.Book{
			StoreId:   storeId,
			CreatedAt: time.Now(),
			Status:    entity.BookStatusCreated,
			Username:  dto.Username,
			Phone:     dto.Phone,
			Message:   dto.Message,
		},
		Products: dto.Products,
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
		BookWithProducts: *book,
		Delay:            int(math.Round(s.bookingDelay.Hours())),
	}, nil
}

func (s *BookingService) GetByIds(ctx context.Context, bookIds []int) ([]GetBookingResponseDTO, error) {
	bookings, err := s.repo.GetByIds(ctx, bookIds)
	if err != nil {
		return nil, err
	}
	resp := make([]GetBookingResponseDTO, len(bookings))

	for i := range bookings {
		resp[i] = GetBookingResponseDTO{
			BookWithProducts: bookings[i],
			Delay:            int(math.Round(s.bookingDelay.Hours())),
		}
	}
	return resp, nil
}

func (s *BookingService) GetByStore(ctx context.Context, storeId int) ([]aggregate.BookWithProducts, error) {
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

func (s *BookingService) StartAutoCancel(ctx context.Context) {
	const delay = time.Minute * 15

	l := contextx.GetLoggerOrDefault(ctx)

	ticker := time.NewTicker(delay)
	for {

		select {
		case <-ticker.C:
		case <-ctx.Done():
			l.Info("stop booking auto cancel", logx.Error(ctx.Err()))
			return
		}

		func() {
			ctx, cancel := context.WithTimeout(context.Background(), delay)
			defer cancel()

			bookings, err := s.repo.GetActive(ctx)
			if err != nil {
				l.Error("failed to get active bookings", logx.Error(err))
				return
			}

			for _, booking := range bookings {
				if booking.CreatedAt.Add(s.bookingDelay).Before(time.Now()) {
					if err := s.repo.UpdateStatus(ctx, booking.Id, entity.BookStatusRejected); err != nil {
						l.Error("cancel booking failed", logx.Error(err), slog.Int("book_id", booking.Id))
						continue
					}
				}
			}
		}()
	}

}
