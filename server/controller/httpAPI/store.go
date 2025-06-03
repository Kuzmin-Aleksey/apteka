package httpAPI

import (
	"encoding/json"
	"net/http"
	"server/internal/domain/entity"
	"server/pkg/failure"
	"strconv"
)

func (h *Handler) ApiNewStore(w http.ResponseWriter, r *http.Request) {
	store := new(entity.Store)
	if err := json.NewDecoder(r.Body).Decode(store); err != nil {
		h.writeError(w, failure.NewInvalidRequestError("invalid json: "+err.Error()))
		return
	}

	if err := h.store.NewStore(r.Context(), store); err != nil {
		h.writeError(w, err)
		return
	}
}

func (h *Handler) ApiGetStores(w http.ResponseWriter, r *http.Request) {
	stores, err := h.store.GetAll(r.Context())
	if err != nil {
		h.writeError(w, err)
		return
	}
	h.writeJSON(w, stores)
}

func (h *Handler) ApiUpdateStore(w http.ResponseWriter, r *http.Request) {
	store := new(entity.Store)
	if err := json.NewDecoder(r.Body).Decode(store); err != nil {
		h.writeError(w, failure.NewInvalidRequestError("invalid json: "+err.Error()))
		return
	}

	if err := h.store.UpdateStore(r.Context(), store); err != nil {
		h.writeError(w, err)
		return
	}
}

func (h *Handler) ApiDeleteStore(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.writeError(w, failure.NewInvalidRequestError("invalid form: "+err.Error()))
		return
	}
	storeId, err := strconv.Atoi(r.Form.Get("store_id"))
	if err != nil {
		h.writeError(w, failure.NewInvalidRequestError("invalid store id: "+err.Error()))
		return
	}

	if err := h.store.DeleteStoreAndProducts(r.Context(), storeId); err != nil {
		h.writeError(w, err)
		return
	}
}
