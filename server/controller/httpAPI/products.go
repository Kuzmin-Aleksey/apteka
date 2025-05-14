package httpAPI

import (
	"net/http"
	"server/domain/models"
	"server/domain/service/products/product_decoder"
	"strconv"
)

func (h *Handler) ApiHandleSearchProducts(w http.ResponseWriter, r *http.Request) {
	storeId, err := strconv.Atoi(r.FormValue("store_id"))
	if err != nil {
		h.writeError(w, models.NewError(models.ErrInvalidRequest, "invalid form value store_id", r.FormValue("store_id"), err))
		return
	}

	q := r.FormValue("q")

	products, err := h.products.Search(r.Context(), storeId, q)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, products)
}

func (h *Handler) ApiHandleUploadProducts(w http.ResponseWriter, r *http.Request) {
	products, err := product_decoder.Decode(r.Body)
	if err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.products.UploadProducts(r.Context(), products); err != nil {
		h.writeError(w, err)
		return
	}

	// go h.images.LoadImages(products)
}
