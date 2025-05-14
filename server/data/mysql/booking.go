package mysql

import (
	"database/sql"
	"errors"
	"golang.org/x/net/context"
	"server/domain/models"
)

type BookingRepo struct {
	DB
}

func NewBookingRepo(db DB) *BookingRepo {
	return &BookingRepo{
		DB: db,
	}
}

func (r *BookingRepo) Save(ctx context.Context, book *models.Book) error {
	tx, err := r.Begin()
	if err != nil {
		return models.NewError(models.ErrDatabaseError, err)
	}

	res, err := r.ExecContext(ctx, "INSERT INTO booking (store_id, created_at, status, username, phone, message) VALUES (?, ?, ?, ?, ?, ?)",
		book.StoreId, book.CreatedAt, book.Status, book.Username, book.Phone, book.Message)
	if err != nil {
		return models.NewError(models.ErrDatabaseError, err)
	}

	bookId, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return models.NewError(models.ErrDatabaseError, err)
	}
	book.Id = int(bookId)

	for _, product := range book.Products {
		if _, err := r.ExecContext(ctx, "INSERT INTO booking_products (booking_id, code_stu, name, quantity) VALUES (?, ?, ?, ?)",
			book.Id, product.CodeSTU, product.Name, product.Quantity); err != nil {
			tx.Rollback()
			return models.NewError(models.ErrDatabaseError, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return models.NewError(models.ErrDatabaseError, err)
	}
	return nil
}

func (r *BookingRepo) UpdateStatus(ctx context.Context, bookId int, status string) error {
	if _, err := r.ExecContext(ctx, `UPDATE booking SET status=? WHERE id=?`, status, bookId); err != nil {
		return models.NewError(models.ErrDatabaseError, err)
	}
	return nil
}

func (r *BookingRepo) GetById(ctx context.Context, bookId int) (*models.Book, error) {
	var book models.Book

	if err := r.QueryRowContext(ctx, "SELECT * FROM booking WHERE id=?", bookId).Scan(&book.Id, &book.StoreId, &book.CreatedAt, &book.Status, &book.Username, &book.Phone, &book.Message); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
	}

	products, err := r.getBookingProducts(ctx, bookId)
	if err != nil {
		return nil, err
	}

	book.Products = products

	return &book, nil
}

func (r *BookingRepo) GetByStore(ctx context.Context, storeId int) ([]models.Book, error) {
	var books []models.Book

	rows, err := r.QueryContext(ctx, "SELECT * FROM booking WHERE store_id=?", storeId)
	if err != nil {
		return nil, models.NewError(models.ErrDatabaseError, err)
	}
	defer rows.Close()

	for rows.Next() {
		var book models.Book
		if err := rows.Scan(&book.Id, &book.StoreId, &book.CreatedAt, &book.Status, &book.Username, &book.Phone, &book.Message); err != nil {
			return nil, models.NewError(models.ErrDatabaseError, err)
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

func (r *BookingRepo) GetActiveProductsByStore(ctx context.Context, storeId int) ([]models.BookProduct, error) {
	const query = `
	SELECT booking_products.code_stu,
	       booking_products.name,
	       booking_products.quantity
	FROM booking_products
	INNER JOIN booking ON booking.id = booking_products.booking_id
	WHERE booking.store_id=? AND
	      booking.status NOT IN (?, ?)
	`

	products := make([]models.BookProduct, 0, 1) // not nil

	rows, err := r.QueryContext(ctx, query, storeId, models.BookStatusReceive, models.BookStatusRejected)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return products, nil
		}
		return nil, models.NewError(models.ErrDatabaseError, err)
	}
	defer rows.Close()

	for rows.Next() {
		var product models.BookProduct

		if err := rows.Scan(&product.CodeSTU, &product.Name, &product.Quantity); err != nil {
			return nil, models.NewError(models.ErrDatabaseError, err)
		}
		products = append(products, product)
	}

	return products, nil
}

func (r *BookingRepo) getBookingProducts(ctx context.Context, bookId int) ([]models.BookProduct, error) {
	products := make([]models.BookProduct, 0, 1) // not nil

	rows, err := r.QueryContext(ctx, "SELECT code_stu, name, quantity FROM booking_products WHERE booking_id=?", bookId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return products, nil
		}
		return nil, models.NewError(models.ErrDatabaseError, err)
	}
	defer rows.Close()

	for rows.Next() {
		var product models.BookProduct

		if err := rows.Scan(&product.CodeSTU, &product.Name, &product.Quantity); err != nil {
			return nil, models.NewError(models.ErrDatabaseError, err)
		}
		products = append(products, product)
	}

	return products, nil
}

func (r *BookingRepo) Delete(ctx context.Context, bookId int) error {
	if _, err := r.ExecContext(ctx, "DELETE FROM booking WHERE id=?", bookId); err != nil {
		return models.NewError(models.ErrDatabaseError, err)
	}
	return nil
}
