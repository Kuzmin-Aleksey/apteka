package httpAPI

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"server/internal/domain/entity"
	"server/pkg/failure"
	"strconv"
)

func (h *Handler) ApiHandleGetPromotion(w http.ResponseWriter, r *http.Request) {
	storeId, err := strconv.Atoi(r.FormValue("store_id"))
	if err != nil {
		h.writeError(w, failure.NewInvalidRequestError("invalid form value store_id"+r.FormValue("store_id")+": "+err.Error()))
		return
	}

	promotion, err := h.promotion.GetInStock(r.Context(), storeId)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, promotion)
}

func (h *Handler) ApiHandleGetAllPromotion(w http.ResponseWriter, r *http.Request) {
	promotion, err := h.promotion.GetAll(r.Context())
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, promotion)
}

func (h *Handler) ApiNewPromotion(w http.ResponseWriter, r *http.Request) {
	promo := new(entity.Promotion)
	b, _ := io.ReadAll(r.Body)

	if err := json.NewDecoder(bytes.NewReader(b)).Decode(promo); err != nil {
		h.writeError(w, failure.NewInvalidRequestError("invalid json"+err.Error()))
		return
	}

	if err := h.promotion.New(r.Context(), promo); err != nil {
		h.writeError(w, err)
		return
	}
}

func (h *Handler) ApiUploadPromotion(w http.ResponseWriter, r *http.Request) {
	promotion, err := h.promotion.UploadPromotionDocument(r.Context(), r.Body)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, promotion)
}

func (h *Handler) ApiUpdatePromotion(w http.ResponseWriter, r *http.Request) {
	promo := new(entity.Promotion)
	if err := json.NewDecoder(r.Body).Decode(promo); err != nil {
		h.writeError(w, failure.NewInvalidRequestError("invalid json"+": "+err.Error()))
		return
	}

	if err := h.promotion.Update(r.Context(), promo); err != nil {
		h.writeError(w, err)
		return
	}
}

func (h *Handler) ApiDeleteAllPromotion(w http.ResponseWriter, r *http.Request) {
	if err := h.promotion.DeleteAll(r.Context()); err != nil {
		h.writeError(w, err)
		return
	}
}

func (h *Handler) ApiDeletePromotion(w http.ResponseWriter, r *http.Request) {
	code, err := strconv.Atoi(r.FormValue("product_code"))
	if err != nil {
		h.writeError(w, failure.NewInvalidRequestError("invalid product code"+": "+err.Error()))
		return
	}

	if err := h.promotion.Delete(r.Context(), code); err != nil {
		h.writeError(w, err)
		return
	}
}
