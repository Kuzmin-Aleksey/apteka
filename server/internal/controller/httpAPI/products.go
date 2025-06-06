package httpAPI

import (
	"encoding/json"
	"net/http"
	"server/internal/domain/service/products"
	"server/internal/domain/service/products/product_decoder"
	"server/pkg/failure"
	"strconv"
)

func (h *Handler) ApiHandleSearchProducts(w http.ResponseWriter, r *http.Request) {
	storeId, err := strconv.Atoi(r.FormValue("store_id"))
	if err != nil {
		h.writeError(w, failure.NewInvalidRequestError("invalid form value store_id "+r.FormValue("store_id")+": "+err.Error()))
		return
	}

	q := r.FormValue("q")

	prods, err := h.products.Search(r.Context(), storeId, q)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, prods)
}

func (h *Handler) ApiHandleUploadProducts(w http.ResponseWriter, r *http.Request) {
	prods, err := product_decoder.Decode(r.Body)
	if err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.products.UploadProducts(r.Context(), prods); err != nil {
		h.writeError(w, err)
		return
	}

	// go h.images.LoadImages(products)
}

func (h *Handler) ApiHandleCheckProductsInStock(w http.ResponseWriter, r *http.Request) {
	checkingProducts := new([]products.CheckInStockProduct)
	if err := json.NewDecoder(r.Body).Decode(checkingProducts); err != nil {
		h.writeError(w, failure.NewInvalidRequestError(err.Error()))
		return
	}

	storeId, err := strconv.Atoi(r.FormValue("store_id"))
	if err != nil {
		h.writeError(w, failure.NewInvalidRequestError("invalid form value store_id "+r.FormValue("store_id")+": "+err.Error()))
		return
	}

	result, err := h.products.CheckInStock(r.Context(), storeId, *checkingProducts)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, result)
}
