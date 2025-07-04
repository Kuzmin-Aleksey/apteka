package mysql

import (
	"database/sql"
	"errors"
	"golang.org/x/net/context"
	"server/internal/domain/aggregate"
	"server/internal/domain/entity"
	"server/pkg/failure"
)

type BookingRepo struct {
	DB
}

func NewBookingRepo(db DB) *BookingRepo {
	return &BookingRepo{
		DB: db,
	}
}

func (r *BookingRepo) Save(ctx context.Context, book *aggregate.BookWithProducts) error {
	tx, err := r.Begin()
	if err != nil {
		return failure.NewInternalError(err.Error())
	}

	res, err := tx.ExecContext(ctx, "INSERT INTO booking (store_id, created_at, status, username, phone, message) VALUES (?, ?, ?, ?, ?, ?)",
		book.StoreId, book.CreatedAt, book.Status, book.Username, book.Phone, book.Message)
	if err != nil {
		return failure.NewInternalError(err.Error())
	}

	bookId, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return failure.NewInternalError(err.Error())
	}
	book.Id = int(bookId)

	for _, product := range book.Products {
		if _, err := tx.ExecContext(ctx, "INSERT INTO booking_products (booking_id, code_stu, name, quantity, price) VALUES (?, ?, ?, ?, ?)",
			book.Id, product.CodeSTU, product.Name, product.Quantity, product.Price); err != nil {
			tx.Rollback()
			return failure.NewInternalError(err.Error())
		}
	}

	if err := tx.Commit(); err != nil {
		return failure.NewInternalError(err.Error())
	}
	return nil
}

func (r *BookingRepo) UpdateStatus(ctx context.Context, bookId int, status string) error {
	if _, err := r.ExecContext(ctx, `UPDATE booking SET status=? WHERE id=?`, status, bookId); err != nil {
		return failure.NewInternalError(err.Error())
	}
	return nil
}

func (r *BookingRepo) GetById(ctx context.Context, bookId int) (*aggregate.BookWithProducts, error) {
	var book aggregate.BookWithProducts

	if err := r.QueryRowContext(ctx, "SELECT * FROM booking WHERE id=?", bookId).Scan(&book.Id, &book.StoreId, &book.CreatedAt, &book.Status, &book.Username, &book.Phone, &book.Message); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, failure.NewNotFoundError(err.Error())
		}
	}

	products, err := r.getBookingProducts(ctx, bookId)
	if err != nil {
		return nil, err
	}

	book.Products = products

	return &book, nil
}

func (r *BookingRepo) GetByIds(ctx context.Context, ids []int) ([]aggregate.BookWithProducts, error) {
	books := make([]aggregate.BookWithProducts, 0, len(ids))

	if len(ids) == 0 {
		return books, nil
	}

	query := "SELECT * FROM booking WHERE id IN (" + joinNums(ids, ", ") + ")"
	rows, err := r.QueryContext(ctx, query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return books, nil
		}
		return nil, failure.NewInternalError(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var book aggregate.BookWithProducts
		if err := rows.Scan(&book.Id, &book.StoreId, &book.CreatedAt, &book.Status, &book.Username, &book.Phone, &book.Message); err != nil {
			return nil, failure.NewInternalError(err.Error())
		}

		products, err := r.getBookingProducts(ctx, book.Id)
		if err != nil {
			return nil, err
		}
		book.Products = products

		books = append(books, book)

	}

	return books, nil
}

func (r *BookingRepo) GetByStore(ctx context.Context, storeId int) ([]aggregate.BookWithProducts, error) {
	var books []aggregate.BookWithProducts

	rows, err := r.QueryContext(ctx, "SELECT * FROM booking WHERE store_id=?", storeId)
	if err != nil {
		return nil, failure.NewInternalError(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var book aggregate.BookWithProducts
		if err := rows.Scan(&book.Id, &book.StoreId, &book.CreatedAt, &book.Status, &book.Username, &book.Phone, &book.Message); err != nil {
			return nil, failure.NewInternalError(err.Error())
		}

		products, err := r.getBookingProducts(ctx, book.Id)
		if err != nil {
			return nil, err
		}
		book.Products = products

		books = append(books, book)

	}

	return books, nil
}

func (r *BookingRepo) GetActive(ctx context.Context) ([]entity.Book, error) {
	books := make([]entity.Book, 0)
	rows, err := r.QueryContext(ctx, "SELECT * FROM booking WHERE booking.status NOT IN (?, ?)", entity.BookStatusReceive, entity.BookStatusRejected)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return books, nil
		}
		return nil, failure.NewInternalError(err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var book entity.Book
		if err := rows.Scan(&book.Id, &book.StoreId, &book.CreatedAt, &book.Status, &book.Username, &book.Phone, &book.Message); err != nil {
			return nil, failure.NewInternalError(err.Error())
		}
		books = append(books, book)
	}
	return books, nil
}

func (r *BookingRepo) GetActiveProductsByStore(ctx context.Context, storeId int) ([]entity.BookProduct, error) {
	const query = `
	SELECT booking_products.code_stu,
	       booking_products.name,
	       booking_products.quantity,
		   booking_products.price
	FROM booking_products
	INNER JOIN booking ON booking.id = booking_products.booking_id
	WHERE booking.store_id=? AND
	      booking.status NOT IN (?, ?)
	`

	products := make([]entity.BookProduct, 0, 1) // not nil

	rows, err := r.QueryContext(ctx, query, storeId, entity.BookStatusReceive, entity.BookStatusRejected)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return products, nil
		}
		return nil, failure.NewInternalError(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var product entity.BookProduct

		if err := rows.Scan(&product.CodeSTU, &product.Name, &product.Quantity, &product.Price); err != nil {
			return nil, failure.NewInternalError(err.Error())
		}
		products = append(products, product)
	}

	return products, nil
}

func (r *BookingRepo) getBookingProducts(ctx context.Context, bookId int) ([]entity.BookProduct, error) {
	products := make([]entity.BookProduct, 0, 1) // not nil

	rows, err := r.QueryContext(ctx, "SELECT code_stu, name, quantity, price FROM booking_products WHERE booking_id=?", bookId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return products, nil
		}
		return nil, failure.NewInternalError(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var product entity.BookProduct

		if err := rows.Scan(&product.CodeSTU, &product.Name, &product.Quantity, &product.Price); err != nil {
			return nil, failure.NewInternalError(err.Error())
		}
		products = append(products, product)
	}

	return products, nil
}

func (r *BookingRepo) Delete(ctx context.Context, bookId int) error {
	if _, err := r.ExecContext(ctx, "DELETE FROM booking WHERE id=?", bookId); err != nil {
		return failure.NewInternalError(err.Error())
	}
	return nil
}
