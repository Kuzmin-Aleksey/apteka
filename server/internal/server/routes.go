package server

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"server/internal/config"
)

const (
	get  = http.MethodGet
	post = http.MethodPost
)

func (s *Server) InitRoutes(rtr *mux.Router, webCfg *config.WebConfig) {
	rtr.HandleFunc("/ping", s.pingHandler)
	rtr.PathPrefix("/image/").Handler(http.StripPrefix("/image", http.FileServer(http.FS(s.imagesFS))))

	s.InitApiHandlers(rtr)
	s.InitRootStaticFilesRoutes(rtr)
	s.InitPagesRoutes(rtr, webCfg)

}

func (s *Server) InitApiHandlers(rtr *mux.Router) {
	rtr.HandleFunc("/api/auth/login", s.auth.ApiHandleLogin).Methods(post)
	rtr.HandleFunc("/api/auth/check-token", s.auth.MwAuth(func(w http.ResponseWriter, r *http.Request) {}))
	rtr.HandleFunc("/api/auth/logout", s.auth.ApiHandleLogout).Methods(post)

	rtr.HandleFunc("/api/products/search", s.products.ApiHandleSearchProducts).Methods(get)
	rtr.HandleFunc("/api/products/check-in-stock", s.products.ApiHandleCheckProductsInStock).Methods(post)
	rtr.Handle("/api/products/upload", s.mwWithApiKey(http.HandlerFunc(s.products.ApiHandleUploadProducts))).Methods(post)

	rtr.HandleFunc("/api/promotion/get", s.promotion.ApiHandleGetPromotion).Methods(get)
	rtr.HandleFunc("/api/promotion/get-all", s.promotion.ApiHandleGetAllPromotion).Methods(get)
	rtr.HandleFunc("/api/promotion/new", s.auth.MwAuth(s.promotion.ApiNewPromotion)).Methods(post)
	rtr.HandleFunc("/api/promotion/upload-file", s.auth.MwAuth(s.promotion.ApiUploadPromotion)).Methods(post)
	rtr.HandleFunc("/api/promotion/update", s.auth.MwAuth(s.promotion.ApiUpdatePromotion)).Methods(post)
	rtr.HandleFunc("/api/promotion/delete", s.auth.MwAuth(s.promotion.ApiDeletePromotion)).Methods(post)
	rtr.HandleFunc("/api/promotion/delete-all", s.auth.MwAuth(s.promotion.ApiDeleteAllPromotion)).Methods(post)

	rtr.HandleFunc("/api/images/load", s.auth.MwAuth(s.images.ApiLoadImages)).Methods(post)
	rtr.HandleFunc("/api/images/load/progress", s.images.ApiLoadImagesProgress) // ws
	rtr.HandleFunc("/api/images/load/stop", s.auth.MwAuth(s.images.ApiImagesStopLoading)).Methods(post)
	rtr.HandleFunc("/api/images/save", s.auth.MwAuth(s.images.ApiSaveImage)).Methods(post)
	rtr.HandleFunc("/api/images/stat", s.auth.MwAuth(s.images.ApiGetImagesStat)).Methods(get)
	rtr.HandleFunc("/api/images/exist", s.images.ApiCheckImage).Methods(get)
	rtr.HandleFunc("/api/images/delete", s.images.ApiDeleteImage).Methods(post)

	rtr.HandleFunc("/api/stores/create", s.auth.MwAuth(s.store.ApiNewStore)).Methods(post)
	rtr.HandleFunc("/api/stores/get", s.store.ApiGetStores).Methods(get)
	rtr.HandleFunc("/api/stores/update", s.auth.MwAuth(s.store.ApiUpdateStore)).Methods(post)
	rtr.HandleFunc("/api/stores/delete", s.auth.MwAuth(s.store.ApiDeleteStore)).Methods(post)

	rtr.HandleFunc("/api/booking/create", s.booking.ApiToBook).Methods(post)
	rtr.HandleFunc("/api/booking/get", s.booking.ApiGetBook).Methods(get)
	rtr.HandleFunc("/api/booking/get-by-ids", s.booking.ApiGetBooks).Methods(get)
	rtr.HandleFunc("/api/booking/get-by-ids/ws", s.booking.ApiSubscribeIdsOnBookingsUpdate).Methods(get)
	rtr.Handle("/api/booking/update-status", s.mwWithApiKey(http.HandlerFunc(s.booking.ApiBookUpdateStatus))).Methods(post)
	rtr.Handle("/api/booking/by-store", s.mwWithApiKey(http.HandlerFunc(s.booking.ApiGetStoreBookings))).Methods(get)
	rtr.Handle("/api/booking/by-store/ws", s.mwWithApiKey(http.HandlerFunc(s.booking.ApiSubscribeStoreOnBookingsUpdate))).Methods(get)
	rtr.Handle("/api/booking/delete", s.mwWithApiKey(http.HandlerFunc(s.booking.ApiDeleteBooking))).Methods(post)
	rtr.HandleFunc("/api/booking/set-delay", s.auth.MwAuth(s.booking.ApiSetBookingDelay)).Methods(post)
	rtr.HandleFunc("/api/booking/get-delay", s.booking.ApiGetBookingDelay).Methods(get)
}

func (s *Server) InitPagesRoutes(rtr *mux.Router, webCfg *config.WebConfig) {
	rtr.PathPrefix("/static/").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("./web/static/"))))
	rtr.HandleFunc("/sitemap.xml", s.sitemapGenerator.handleSitemap)

	tm := newTemplateHandler(webCfg)

	rtr.Handle("/", tm.handleTemplate("web/templates/main_page.html", "web/templates/navbar.html", "web/templates/booking-cart.html", "web/templates/footer.html"))
	rtr.Handle("/stores", tm.handleTemplate("web/templates/stores_page.html", "web/templates/navbar.html", "web/templates/booking-cart.html", "web/templates/footer.html"))
	rtr.Handle("/bookings", tm.handleTemplate("web/templates/bookings_page.html", "web/templates/navbar.html", "web/templates/booking-cart.html", "web/templates/footer.html"))
	rtr.Handle("/admin/login", tm.handleTemplate("web/templates/login.html"))
	rtr.Handle("/admin", tm.handleTemplate("web/templates/admin.html", "web/templates/admin_root.html"))
	rtr.Handle("/admin/stores", tm.handleTemplate("web/templates/admin.html", "web/templates/admin_store.html"))
	rtr.Handle("/admin/promotion", tm.handleTemplate("web/templates/admin.html", "web/templates/admin_promotion.html"))
	rtr.Handle("/admin/images", tm.handleTemplate("web/templates/admin.html", "web/templates/admin_images.html"))
}

func (s *Server) InitRootStaticFilesRoutes(rtr *mux.Router) {
	filesInfo, err := os.ReadDir("./web/root")
	if err != nil {
		log.Println("Error opening web root directory:", err)
	}

	for _, fileInfo := range filesInfo {
		if fileInfo.IsDir() {
			continue
		}

		log.Println("handle: ", "/"+fileInfo.Name())

		rtr.Handle("/"+fileInfo.Name(), s.handleFile("web/root/"+fileInfo.Name()))
	}
}

func (s *Server) pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong"))
}
