package app

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/lmittmann/tint"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"server/internal/config"
	"server/internal/domain/service/auth"
	"server/internal/domain/service/booking"
	"server/internal/domain/service/images"
	"server/internal/domain/service/products"
	"server/internal/domain/service/promotion"
	"server/internal/domain/service/store"
	"server/internal/infrastructure/integration/image_parser"
	"server/internal/infrastructure/integration/sphinx"
	"server/internal/infrastructure/persistence/cache/redis"
	"server/internal/infrastructure/persistence/mysql"
	"server/internal/server"
	"server/pkg/contextx"
	"server/pkg/fs"
	"server/pkg/logx"
	"server/pkg/middlewarex"
	"server/pkg/tx_manager"
	"syscall"
	"time"
)

func Run(cfg *config.Config) {
	shutdown := make(chan os.Signal, 2)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	l := initLogger(cfg.Debug)

	db, err := mysql.Connect(cfg.DB)
	if err != nil {
		log.Fatal("connect to mysql fail: ", err)
	}

	defer db.Close()
	txDB := tx_manager.NewDBWithTx(db)

	c, err := cache.Connect(cfg.Redis)
	if err != nil {
		log.Fatal("connect to cache fail: ", err)
	}
	defer c.Close()

	imagesFS, err := fs.NewFS(cfg.Images.FileRoot)
	if err != nil {
		log.Fatal(err)
	}

	sphinxSearcher, err := sphinx.NewSearcher(cfg.Sphinx)
	if err != nil {
		log.Fatal("connect to sphinx fail: ", err)
	}

	productsRepo := mysql.NewProductRepo(txDB)
	promotionRepo := mysql.NewPromotion(txDB)
	storesRepo := mysql.NewStoreRepo(txDB)
	bookingRepo := mysql.NewBookingRepo(txDB)
	imagesParser := image_parser.NewImagesParser(time.Second * 15)

	productsService := products.NewProductService(productsRepo, sphinxSearcher, storesRepo, bookingRepo)
	promotionService := promotion.NewPromotionService(promotionRepo, productsService)
	imagesService := images.NewImagesService(imagesFS, productsRepo, imagesParser)
	storeService := store.NewStoreService(storesRepo, productsRepo)
	authService := auth.NewAuthService(cache.NewTokenCacheAdapter(c), cfg.Auth)
	bookingService, err := booking.NewBookingService(bookingRepo, storesRepo, productsService, promotionRepo)
	if err != nil {
		log.Fatal("init booking service failed: ", err)
	}

	runningCtx, cancelRunning := context.WithCancel(contextx.WithLogger(context.Background(), l))

	if cfg.Promotion.AutoDelete {
		go promotionService.RunAutoDeletion(runningCtx)
	}

	if cfg.Images.AutoLoadDelay >= time.Minute {
		go imagesService.RunAutoDownloader(runningCtx, cfg.Images.AutoLoadDelay)
	}

	go bookingService.StartAutoCancel(runningCtx)

	httpServer := newHttpServer(l, productsService, promotionService, imagesService, storeService, authService, bookingService, imagesFS, cfg.Http, cfg.Web)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	sig := <-shutdown
	log.Println("exit by signal: ", sig)

	cancelRunning()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Println("http server shutdown error: ", err)
	}
	cancel()

	os.Exit(0)
}

func newHttpServer(
	l *slog.Logger,
	productsService *products.ProductService,
	promotionService *promotion.PromotionService,
	imagesService *images.ImagesService,
	storeService *store.StoreService,
	authService *auth.AuthService,
	bookingService *booking.BookingService,
	imagesFS *fs.FS,
	cfg *config.HttpConfig,
	webCfg *config.WebConfig,
) *http.Server {
	productsServer := server.NewProductsServer(productsService)
	promotionServer := server.NewPromotionServer(promotionService)
	imagesServer := server.NewImagesServer(imagesService)
	storeServer := server.NewStoreServer(storeService)
	authServer := server.NewAuthServer(authService)
	bookingServer := server.NewBookingServer(bookingService)
	sitemapGenerator := server.NewSitemapGenerator(storeService)

	s := server.NewServer(
		productsServer,
		promotionServer,
		imagesServer,
		storeServer,
		authServer,
		bookingServer,
		sitemapGenerator,
		imagesFS,
		middlewarex.WithApiKey(cfg.ApiKey),
	)

	server.NewSitemapGenerator(storeService)

	rtr := mux.NewRouter()
	s.InitRoutes(rtr, webCfg)

	rtr.Use(
		middlewarex.TraceId,
		middlewarex.Logger,
		middlewarex.RequestLogging(logx.NewSensitiveDataMasker(), 3000),
		middlewarex.ResponseLogging(logx.NewSensitiveDataMasker(), 3000),
		middlewarex.NoCache,
		middlewarex.Recovery,
	)
	if cfg.HandleTimeoutSec > 0 {
		rtr.Use(middlewarex.WithTimeout(time.Duration(cfg.HandleTimeoutSec) * time.Second))
	}

	return &http.Server{
		Addr:         cfg.Address,
		Handler:      rtr,
		ReadTimeout:  time.Duration(cfg.ReadTimeoutSec) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeoutSec) * time.Second,
		ErrorLog:     slog.NewLogLogger(l.Handler(), slog.LevelError),
		BaseContext: func(net.Listener) context.Context {
			return contextx.WithLogger(context.Background(), l)
		},
	}
}

func initLogger(debug bool) *slog.Logger {
	if debug {
		return slog.New(tint.NewHandler(os.Stdout, &tint.Options{
			AddSource:   false,
			Level:       slog.LevelDebug,
			ReplaceAttr: nil,
			TimeFormat:  time.StampMilli,
			NoColor:     false,
		}))
	}

	f, err := os.OpenFile("logs.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	return slog.New(slog.NewJSONHandler(io.MultiWriter(os.Stdout, f), nil))
}
