package httpAPI

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"server/domain/models"
	"strconv"
)

func (h *Handler) ApiHandleGetPromotion(w http.ResponseWriter, r *http.Request) {
	storeId, err := strconv.Atoi(r.FormValue("store_id"))
	if err != nil {
		h.writeError(w, models.NewError(models.ErrInvalidRequest, "invalid form value store_id", r.FormValue("store_id"), err))
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
	promo := new(models.Promotion)
	b, _ := io.ReadAll(r.Body)

	if err := json.NewDecoder(bytes.NewReader(b)).Decode(promo); err != nil {
		h.writeError(w, models.NewError(models.ErrInvalidRequest, "invalid json", err))
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
	promo := new(models.Promotion)
	if err := json.NewDecoder(r.Body).Decode(promo); err != nil {
		h.writeError(w, models.NewError(models.ErrInvalidRequest, "invalid json", err))
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
		h.writeError(w, models.NewError(models.ErrInvalidRequest, "invalid product code", err))
		return
	}

	if err := h.promotion.Delete(r.Context(), code); err != nil {
		h.writeError(w, err)
		return
	}
}
