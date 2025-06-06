package mysql

import (
	"database/sql"
	"errors"
	"golang.org/x/net/context"
	"server/internal/domain/entity"
	"server/pkg/failure"
	"time"
)

type StoreRepo struct {
	DB
}

func NewStoreRepo(db DB) *StoreRepo {
	return &StoreRepo{
		DB: db,
	}
}

func (r *StoreRepo) New(ctx context.Context, store *entity.Store) error {
	res, err := r.DB.ExecContext(ctx, "INSERT INTO stores (address, pos_lat, pos_lon, mobile, email, booking_enable, schedule) VALUES (?, ?, ?, ?, ?, ?, ?)",
		store.Address, store.Position.Lat, store.Position.Lon, store.Contacts.Mobile, store.Contacts.Email, store.BookingEnable, store.Schedule)
	if err != nil {
		return failure.NewInternalError(err.Error())
	}
	id, _ := res.LastInsertId()
	store.Id = int(id)
	return nil
}

func (r *StoreRepo) Update(ctx context.Context, store *entity.Store) error {
	if _, err := r.DB.ExecContext(ctx, "UPDATE stores SET address=?, pos_lat=?, pos_lon=?, mobile=?, email=?, booking_enable=?, schedule=? WHERE id=? ",
		store.Address, store.Position.Lat, store.Position.Lon, store.Contacts.Mobile, store.Contacts.Email, store.BookingEnable, store.Schedule, store.Id); err != nil {
		return failure.NewInternalError(err.Error())
	}
	return nil
}

func (r *StoreRepo) SetLastUploadTime(ctx context.Context, storeId int, t time.Time) error {
	if _, err := r.DB.ExecContext(ctx, "UPDATE stores SET upload_time=? WHERE id=? ", t, storeId); err != nil {
		return failure.NewInternalError(err.Error())
	}
	return nil
}

func (r *StoreRepo) Delete(ctx context.Context, id int) error {
	if _, err := r.DB.ExecContext(ctx, "DELETE FROM stores WHERE id=?", id); err != nil {
		return failure.NewInternalError(err.Error())
	}
	return nil
}

func (r *StoreRepo) GetAll(ctx context.Context) ([]entity.Store, error) {
	var stores []entity.Store
	rows, err := r.DB.QueryContext(ctx, "SELECT * FROM stores")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return stores, nil
		}
		return nil, failure.NewInternalError(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var store entity.Store
		if err := rows.Scan(&store.Id, &store.Address, &store.UploadTime, &store.Position.Lat, &store.Position.Lon, &store.Contacts.Mobile, &store.Contacts.Email, &store.BookingEnable, &store.Schedule); err != nil {
			return nil, failure.NewInternalError(err.Error())
		}
		stores = append(stores, store)
	}
	return stores, nil
}

func (r *StoreRepo) IsBookingEnable(ctx context.Context, id int) (bool, error) {
	var enable bool
	if err := r.DB.QueryRowContext(ctx, "SELECT booking_enable FROM stores WHERE id=?", id).Scan(&enable); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, failure.NewInternalError(err.Error())
	}
	return enable, nil
}
