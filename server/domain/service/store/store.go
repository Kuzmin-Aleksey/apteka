package store

import (
	"golang.org/x/net/context"
	"server/domain/models"
)

type ProductsRepo interface {
	DeleteByStoreId(ctx context.Context, storeID int) error
}

type StoreRepo interface {
	New(ctx context.Context, store *models.Store) error
	Update(ctx context.Context, store *models.Store) error
	Delete(ctx context.Context, id int) error
	GetAll(ctx context.Context) ([]models.Store, error)
}

type StoreService struct {
	storeRepo    StoreRepo
	productsRepo ProductsRepo
}

func NewStoreService(storeRepo StoreRepo, productsRepo ProductsRepo) *StoreService {
	return &StoreService{
		storeRepo:    storeRepo,
		productsRepo: productsRepo,
	}
}

func (s *StoreService) NewStore(ctx context.Context, store *models.Store) error {
	if err := s.storeRepo.New(ctx, store); err != nil {
		return err
	}
	return nil
}

func (s *StoreService) UpdateStore(ctx context.Context, store *models.Store) error {
	if err := s.storeRepo.Update(ctx, store); err != nil {
		return err
	}
	return nil
}

func (s *StoreService) DeleteStoreAndProducts(ctx context.Context, storeId int) error {
	if err := s.storeRepo.Delete(ctx, storeId); err != nil {
		return err
	}
	if err := s.productsRepo.DeleteByStoreId(ctx, storeId); err != nil {
		return err
	}
	return nil
}

func (s *StoreService) GetAll(ctx context.Context) ([]models.Store, error) {
	stores, err := s.storeRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return stores, nil
}
