package mysql

import (
	"database/sql"
	"errors"
	"golang.org/x/net/context"
	"server/domain/models"
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

func (r *StoreRepo) New(ctx context.Context, store *models.Store) error {
	res, err := r.DB.ExecContext(ctx, "INSERT INTO stores (address, upload_time, pos_lat, pos_lon, mobile, email) VALUES (?, ?, ?, ?, ?, ?)",
		store.Address, store.UploadTime, store.Position.Lat, store.Position.Lon, store.Contacts.Mobile, store.Contacts.Email)
	if err != nil {
		return models.NewError(models.ErrDatabaseError, err)
	}
	id, _ := res.LastInsertId()
	store.Id = int(id)
	return nil
}

func (r *StoreRepo) Update(ctx context.Context, store *models.Store) error {
	if _, err := r.DB.ExecContext(ctx, "UPDATE stores SET address=?, upload_time=?, pos_lat=?, pos_lon=?, mobile=?, email=? WHERE id=? ",
		store.Address, store.UploadTime, store.Position.Lat, store.Position.Lon, store.Contacts.Mobile, store.Contacts.Email, store.Id); err != nil {
		return models.NewError(models.ErrDatabaseError, err)
	}
	return nil
}

func (r *StoreRepo) SetLastUploadTime(ctx context.Context, storeId int, t time.Time) error {
	if _, err := r.DB.ExecContext(ctx, "UPDATE stores SET upload_time=? WHERE id=? ", t, storeId); err != nil {
		return err
	}
	return nil
}

func (r *StoreRepo) Delete(ctx context.Context, id int) error {
	if _, err := r.DB.ExecContext(ctx, "DELETE FROM stores WHERE id=?", id); err != nil {
		return models.NewError(models.ErrDatabaseError, err)
	}
	return nil
}

func (r *StoreRepo) GetAll(ctx context.Context) ([]models.Store, error) {
	var stores []models.Store
	rows, err := r.DB.QueryContext(ctx, "SELECT * FROM stores")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return stores, nil
		}
		return nil, models.NewError(models.ErrDatabaseError, err)
	}
	defer rows.Close()

	for rows.Next() {
		var store models.Store
		if err := rows.Scan(&store.Id, &store.Address, &store.UploadTime, &store.Position.Lat, &store.Position.Lon, &store.Contacts.Mobile, &store.Contacts.Email); err != nil {
			return nil, models.NewError(models.ErrDatabaseError, err)
		}
		stores = append(stores, store)
	}
	return stores, nil
}
