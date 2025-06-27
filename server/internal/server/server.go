package server

import (
	"io/fs"
	"net/http"
)

type MiddleWareFunc func(next http.Handler) http.Handler

type Server struct {
	products         *ProductsServer
	promotion        *PromotionServer
	images           *ImagesServer
	store            *StoreServer
	auth             *AuthServer
	booking          *BookingServer
	sitemapGenerator *SitemapGenerator

	imagesFS fs.FS

	mwWithApiKey MiddleWareFunc
}

func NewServer(
	products *ProductsServer,
	promotion *PromotionServer,
	images *ImagesServer,
	store *StoreServer,
	auth *AuthServer,
	booking *BookingServer,
	sitemapGenerator *SitemapGenerator,
	imagesFS fs.FS,
	mwWithApiKey MiddleWareFunc,
) *Server {
	return &Server{
		products:         products,
		promotion:        promotion,
		images:           images,
		store:            store,
		auth:             auth,
		booking:          booking,
		sitemapGenerator: sitemapGenerator,
		imagesFS:         imagesFS,
		mwWithApiKey:     mwWithApiKey,
	}
}
