package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"server/internal/domain/entity"
	"server/internal/domain/service/promotion"
	"server/pkg/failure"
	"strconv"
)

type PromotionServer struct {
	promotion *promotion.PromotionService
}

func NewPromotionServer(promotion *promotion.PromotionService) *PromotionServer {
	return &PromotionServer{promotion: promotion}
}

func (s *PromotionServer) ApiHandleGetPromotion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	storeId, err := strconv.Atoi(r.FormValue("store_id"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid form value store_id"+r.FormValue("store_id")+": "+err.Error()))
		return
	}

	promo, err := s.promotion.GetInStock(ctx, storeId)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}

	writeJson(ctx, w, promo, http.StatusOK)
}

func (s *PromotionServer) ApiHandleGetAllPromotion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	promo, err := s.promotion.GetAll(ctx)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}

	writeJson(ctx, w, promo, http.StatusOK)
}

func (s *PromotionServer) ApiNewPromotion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	promo := new(entity.Promotion)
	b, _ := io.ReadAll(r.Body)

	if err := json.NewDecoder(bytes.NewReader(b)).Decode(promo); err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid json"+err.Error()))
		return
	}

	if err := s.promotion.New(ctx, promo); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
}

func (s *PromotionServer) ApiUploadPromotion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	promo, err := s.promotion.UploadPromotionDocument(ctx, r.Body)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}

	writeJson(ctx, w, promo, http.StatusOK)
}

func (s *PromotionServer) ApiUpdatePromotion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	promo := new(entity.Promotion)
	if err := json.NewDecoder(r.Body).Decode(promo); err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid json"+": "+err.Error()))
		return
	}

	if err := s.promotion.Update(r.Context(), promo); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
}

func (s *PromotionServer) ApiDeleteAllPromotion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.promotion.DeleteAll(ctx); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
}

func (s *PromotionServer) ApiDeletePromotion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	code, err := strconv.Atoi(r.FormValue("product_code"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid product code"+": "+err.Error()))
		return
	}

	if err := s.promotion.Delete(ctx, code); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
}
