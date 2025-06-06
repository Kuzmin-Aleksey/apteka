package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"server/config"
	"server/internal/app/app/server"
	"server/internal/controller/httpAPI"
	"server/internal/domain/service/auth"
	"server/internal/domain/service/booking"
	"server/internal/domain/service/images"
	"server/internal/domain/service/products"
	"server/internal/domain/service/promotion"
	"server/internal/domain/service/store"
	"server/internal/infrastructure/integration/image_parser"
	"server/internal/infrastructure/persistence/cache/redis"
	"server/internal/infrastructure/persistence/mysql"
	"server/pkg/fs"
	"server/pkg/logger"
	"server/pkg/tx_manager"
	"syscall"
	"time"
)

func Run(cfg *config.Config) {
	shutdown := make(chan os.Signal, 2)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	l, err := logger.NewLogger()
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	db, err := mysql.Connect(cfg.DB)
	if err != nil {
		l.Fatal("connect to mysql fail: ", err)
	}

	defer db.Close()
	txDB := tx_manager.NewDBWithTx(db)

	c, err := cache.Connect(cfg.Redis)
	if err != nil {
		l.Fatal("connect to cache fail: ", err)
	}
	defer c.Close()

	imagesFS, err := fs.NewFS(cfg.Images.FileRoot)
	if err != nil {
		l.Fatal(err)
	}

	productsRepo := mysql.NewProductRepo(txDB)
	promotionRepo := mysql.NewPromotion(txDB)
	storesRepo := mysql.NewStoreRepo(txDB)
	bookingRepo := mysql.NewBookingRepo(txDB)
	imagesParser := image_parser.NewImagesParser(time.Second * 15)

	productsService := products.NewProductService(productsRepo, storesRepo, bookingRepo)
	promotionService := promotion.NewPromotionService(promotionRepo, productsService, l)
	imagesService := images.NewImagesService(imagesFS, productsRepo, imagesParser, l)
	storeService := store.NewStoreService(storesRepo, productsRepo)
	authService := auth.NewAuthService(cache.NewTokenCacheAdapter(c), cfg.Auth)
	bookingService, err := booking.NewBookingService(bookingRepo, storesRepo, productsService, promotionRepo)
	if err != nil {
		l.Fatal(err)
	}

	if cfg.Promotion.AutoDelete {
		go promotionService.RunAutoDeletion()
	}

	if cfg.Images.AutoLoadDelay >= time.Minute {
		go imagesService.RunAutoDownloader(cfg.Images.AutoLoadDelay)
	}

	handler, err := httpAPI.NewHandler(
		productsService,
		promotionService,
		imagesService,
		storeService,
		authService,
		bookingService,
		imagesFS,
		l,
		cfg.Api,
	)
	if err != nil {
		l.Fatal(err)
	}

	srv := server.CreateHttpServer(handler, l, cfg.Http)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			l.Fatal(err)
		}
	}()

	sig := <-shutdown
	l.Println("exit by signal: ", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := srv.Shutdown(ctx); err != nil {
		l.Println("http server shutdown error: ", err)
	}
	cancel()

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	if err := promotionService.Shutdown(ctx); err != nil {
		l.Println("shutdown promotion service error: ", err)
	}
	cancel()

	os.Exit(0)
}
