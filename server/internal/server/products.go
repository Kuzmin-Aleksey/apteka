package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"server/internal/domain/service/products"
	"server/internal/domain/service/products/product_decoder"
	"server/pkg/failure"
	"strconv"
)

type ProductsServer struct {
	products *products.ProductService
}

func NewProductsServer(products *products.ProductService) *ProductsServer {
	return &ProductsServer{products: products}
}

func (s *ProductsServer) ApiHandleSearchProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	storeId, err := strconv.Atoi(r.FormValue("store_id"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid form value store_id "+r.FormValue("store_id")+": "+err.Error()))
		return
	}

	q := r.FormValue("q")

	prods, err := s.products.Search(r.Context(), storeId, q)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}

	writeJson(ctx, w, prods, http.StatusOK)
}

func (s *ProductsServer) ApiHandleUploadProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(err.Error()))
		return
	}

	prods, storeId, err := product_decoder.Decode(bytes.NewReader(body))
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}

	if err := s.products.UploadProducts(r.Context(), storeId, prods); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
}

func (s *ProductsServer) ApiHandleCheckProductsInStock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	checkingProducts := new([]products.CheckInStockProduct)
	if err := json.NewDecoder(r.Body).Decode(checkingProducts); err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(err.Error()))
		return
	}

	storeId, err := strconv.Atoi(r.FormValue("store_id"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid form value store_id "+r.FormValue("store_id")+": "+err.Error()))
		return
	}

	result, err := s.products.CheckInStock(r.Context(), storeId, *checkingProducts)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}

	writeJson(ctx, w, result, http.StatusOK)
}
