package httpAPI

import (
	"fmt"
	"github.com/gorilla/mux"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"server/config"
	"server/domain/service/auth"
	"server/domain/service/booking"
	"server/domain/service/images"
	"server/domain/service/products"
	"server/domain/service/promotion"
	"server/domain/service/store"
	"time"
)

type Logger interface {
	Println(v ...any)
	Printf(format string, v ...any)
}

type Handler struct {
	router *mux.Router

	products  *products.ProductService
	promotion *promotion.PromotionService
	images    *images.ImagesService
	store     *store.StoreService
	auth      *auth.AuthService
	booking   *booking.BookingService
	imagesFS  fs.FS

	info *HttpLogger
	l    Logger

	ApiKey  string
	Timeout time.Duration
}

func NewHandler(
	products *products.ProductService,
	promotion *promotion.PromotionService,
	images *images.ImagesService,
	store *store.StoreService,
	auth *auth.AuthService,
	booking *booking.BookingService,
	imagesFS fs.FS,

	l Logger,
	cfg *config.ApiConfig,

) (*Handler, error) {
	requestsLogs, err := os.OpenFile(filepath.Join("logs", "requests.log.csv"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("create requests log file error: %w", err)
	}

	h := &Handler{
		products:  products,
		promotion: promotion,
		images:    images,
		store:     store,
		auth:      auth,
		booking:   booking,
		imagesFS:  imagesFS,
		ApiKey:    cfg.ApiKey,
		Timeout:   time.Duration(cfg.HandleTimeoutSec) * time.Second,
		info:      NewHttpLogger(requestsLogs),
		l:         l,
	}

	if err := h.init(); err != nil {
		return nil, err
	}

	return h, nil
}

const (
	get  = http.MethodGet
	post = http.MethodPost
)

func (h *Handler) init() error {
	h.router = mux.NewRouter()

	h.router.NotFoundHandler = h.MwLogging(http.HandlerFunc(h.HandleNotFound))

	h.router.HandleFunc("/ping", h.pingHandler)

	h.router.PathPrefix("/image/").Handler(http.StripPrefix("/image", http.FileServer(http.FS(h.imagesFS))))
	h.router.PathPrefix("/static/").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("./web/static/"))))

	h.router.HandleFunc("/", h.HandleMainPage).Methods(get)
	h.router.HandleFunc("/stores", h.HandleStoresPage).Methods(get)

	adminPagesHandler, err := NewAdminPagesHandler(h)
	if err != nil {
		return err
	}
	h.router.PathPrefix("/admin/").Handler(http.StripPrefix("/admin", adminPagesHandler))
	h.router.Handle("/admin", http.RedirectHandler("/admin/", http.StatusMovedPermanently))

	h.router.PathPrefix("/api/").Handler(http.StripPrefix("/api", h.apiHandler()))

	h.router.Use(
		h.MwWithTimeout,
	)

	return nil
}

func (h *Handler) apiHandler() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/auth/login", h.ApiHandleLogin).Methods(post)
	router.HandleFunc("/auth/check-token", h.MwAuth(func(w http.ResponseWriter, r *http.Request) {}))
	router.HandleFunc("/auth/logout", h.ApiHandleLogout).Methods(post)

	router.HandleFunc("/products/search", h.ApiHandleSearchProducts).Methods(get)
	router.HandleFunc("/products/upload", h.MwWithApiKey(h.ApiHandleUploadProducts)).Methods(post)

	router.HandleFunc("/promotion/get", h.ApiHandleGetPromotion).Methods(get)
	router.HandleFunc("/promotion/get-all", h.ApiHandleGetAllPromotion).Methods(get)
	router.HandleFunc("/promotion/new", h.MwAuth(h.ApiNewPromotion)).Methods(post)
	router.HandleFunc("/promotion/upload-file", h.MwAuth(h.ApiUploadPromotion)).Methods(post)
	router.HandleFunc("/promotion/update", h.MwAuth(h.ApiUpdatePromotion)).Methods(post)
	router.HandleFunc("/promotion/delete", h.MwAuth(h.ApiDeletePromotion)).Methods(post)
	router.HandleFunc("/promotion/delete-all", h.MwAuth(h.ApiDeleteAllPromotion)).Methods(post)

	router.HandleFunc("/images/load", h.MwAuth(h.ApiLoadImages)).Methods(post)
	router.HandleFunc("/images/load/progress", h.ApiLoadImagesProgress) // ws
	router.HandleFunc("/images/load/stop", h.MwAuth(h.ApiImagesStopLoading)).Methods(post)
	router.HandleFunc("/images/save", h.MwAuth(h.ApiSaveImage)).Methods(post)
	router.HandleFunc("/images/stat", h.MwAuth(h.ApiGetImagesStat)).Methods(get)
	router.HandleFunc("/images/exist", h.ApiCheckImage).Methods(get)
	router.HandleFunc("/images/delete", h.ApiDeleteImage).Methods(post)

	router.HandleFunc("/stores/create", h.MwAuth(h.ApiNewStore)).Methods(post)
	router.HandleFunc("/stores/get", h.ApiGetStores).Methods(get)
	router.HandleFunc("/stores/update", h.MwAuth(h.ApiUpdateStore)).Methods(post)
	router.HandleFunc("/stores/delete", h.MwAuth(h.ApiDeleteStore)).Methods(post)

	router.HandleFunc("/booking/create", h.ApiToBook).Methods(post)
	router.HandleFunc("/booking/get", h.ApiGetBook).Methods(get)
	router.HandleFunc("/booking/update-status", h.MwWithApiKey(h.ApiBookUpdateStatus)).Methods(post)
	router.HandleFunc("/booking/by-store", h.MwWithApiKey(h.ApiGetStoreBookings)).Methods(get)
	router.HandleFunc("/booking/delete", h.MwWithApiKey(h.ApiDeleteBooking)).Methods(post)
	router.HandleFunc("/booking/set-delay", h.MwAuth(h.ApiSetBookingDelay)).Methods(post)

	router.Use(
		h.MwNoCache,
		h.MwLogging)

	return router
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *Handler) pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong"))
}
