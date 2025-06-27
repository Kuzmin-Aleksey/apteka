package server

import (
	"encoding/json"
	"net/http"
	"server/internal/domain/entity"
	"server/internal/domain/service/store"
	"server/pkg/failure"
	"strconv"
)

type StoreServer struct {
	store *store.StoreService
}

func NewStoreServer(store *store.StoreService) *StoreServer {
	return &StoreServer{store: store}
}

func (s *StoreServer) ApiNewStore(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	st := new(entity.Store)
	if err := json.NewDecoder(r.Body).Decode(st); err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid json: "+err.Error()))
		return
	}

	if err := s.store.NewStore(ctx, st); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
}

func (s *StoreServer) ApiGetStores(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stores, err := s.store.GetAll(ctx)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
	writeJson(ctx, w, stores, http.StatusOK)
}

func (s *StoreServer) ApiUpdateStore(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	st := new(entity.Store)
	if err := json.NewDecoder(r.Body).Decode(st); err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid json: "+err.Error()))
		return
	}

	if err := s.store.UpdateStore(ctx, st); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
}

func (s *StoreServer) ApiDeleteStore(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	storeId, err := strconv.Atoi(r.FormValue("store_id"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid store id: "+err.Error()))
		return
	}

	if err := s.store.DeleteStoreAndProducts(ctx, storeId); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
}
