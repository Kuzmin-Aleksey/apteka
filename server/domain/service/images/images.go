package images

import (
	"fmt"
	"golang.org/x/net/context"
	"io"
	"log"
	"server/domain/models"
	"sync"
	"time"
)

type ImagesRepo interface {
	Save(data io.Reader, name string) error
	CheckExist(name string) bool
	Count() (int, error)
	Remove(name string) error
}

type ImageParser interface {
	LoadWebpImageByGTIN(ctx context.Context, gtin uint64) (io.ReadCloser, error)
}

type ProductGetter interface {
	GetAll(ctx context.Context) ([]models.Product, error)
}

type Logger interface {
	Printf(format string, v ...any)
}

type ProgressFunc func(int8)

type ImagesService struct {
	repo           ImagesRepo
	products       ProductGetter
	parser         ImageParser
	l              Logger
	lastDownloadAt time.Time
	statCache      *statCache
	IsLoading      bool
	initLoadingMu  sync.Mutex

	stopLoadingCh   chan struct{}
	stopLoadingMu   sync.Mutex
	CurrentProgress int8
}

func NewImagesService(repo ImagesRepo, products ProductGetter, parser ImageParser, l Logger) *ImagesService {
	s := &ImagesService{
		repo:          repo,
		parser:        parser,
		products:      products,
		l:             l,
		stopLoadingCh: make(chan struct{}),
	}

	return s
}

func (s *ImagesService) RunAutoDownloader(delay time.Duration) {
	const op = "ImagesService AutoDownloader"
	time.Sleep(time.Second * 3)

	ticker := time.NewTicker(delay)

	for {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), delay)
			defer cancel()

			if err := s.LoadImages(ctx); err != nil {
				s.l.Printf("%s - %s", op, err)
			}
		}()

		<-ticker.C
	}
}

func (s *ImagesService) CheckImageExist(prodId int) bool {
	return s.repo.CheckExist(getImageName(prodId))
}

func (s *ImagesService) LoadImages(ctx context.Context) error {
	s.initLoadingMu.Lock()
	defer s.initLoadingMu.Unlock()

	if s.IsLoading {
		return nil
	}

	products, err := s.products.GetAll(ctx)
	if err != nil {
		return err
	}
	var productsToLoad []*models.Product

	for _, prod := range products {
		if prod.GTIN == 0 {
			continue
		}
		if s.CheckImageExist(prod.CodeSTU) {
			continue
		}
		productsToLoad = append(productsToLoad, &prod)
	}

	go func() {
		s.IsLoading = true

		var i float64
		var count int

		defer func() {
			s.CurrentProgress = 0
			s.IsLoading = false
			log.Printf("loaded %d images", count)
		}()

		for _, prod := range productsToLoad {
			i++
			select {
			case <-ctx.Done():
				s.l.Printf("load images failed, %s", ctx.Err())
				return
			case <-s.stopLoadingCh:
				s.l.Printf("loading stoped")
				return
			default:
			}

			if err := s.LoadImage(ctx, prod); err != nil {
				s.l.Printf("failed load image for prod (code_stu <%d> gtin <%d> name <%s>): %s", prod.CodeSTU, prod.GTIN, prod.Name, err.Error())
			} else {
				s.l.Printf("loaded image for prod (code_stu <%d> gtin <%d> name <%s>)", prod.CodeSTU, prod.GTIN, prod.Name)
				count++
			}

			s.CurrentProgress = int8(i / float64(len(productsToLoad)) * 100)
		}
	}()

	return nil
}

func (s *ImagesService) StopLoading(ctx context.Context) error {
	s.stopLoadingMu.Lock()
	defer s.stopLoadingMu.Unlock()

	if s.IsLoading {
		select {
		case s.stopLoadingCh <- struct{}{}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (s *ImagesService) LoadImage(ctx context.Context, prod *models.Product) error {
	img, err := s.parser.LoadWebpImageByGTIN(ctx, prod.GTIN)
	if err != nil {
		return fmt.Errorf("load image failed: %w", err)
	}

	if img == nil {
		return fmt.Errorf("returned nil image")
	}

	if err := s.repo.Save(img, getImageName(prod.CodeSTU)); err != nil {
		return fmt.Errorf("save image failed: %w", err)
	}
	return nil
}

func (s *ImagesService) SaveWebpImage(prodCode int, r io.Reader) error {
	if err := s.repo.Save(r, fmt.Sprintf("%d.webp", prodCode)); err != nil {
		return fmt.Errorf("save image failed: %w", err)
	}
	return nil
}

type Stat struct {
	ProductsAll  int `json:"products_all"`
	WithoutImage int `json:"without_image"`
	ImagesAll    int `json:"images_all"`
}

const cacheTTL = time.Second * 5

type statCache struct {
	t    time.Time
	stat *Stat
}

func (s *ImagesService) GetStat(ctx context.Context) (*Stat, error) {
	if s.statCache != nil {
		if time.Since(s.statCache.t) < cacheTTL {
			return s.statCache.stat, nil
		}
	}

	products, err := s.products.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	countImages, err := s.repo.Count()
	if err != nil {
		return nil, err
	}

	var countWithoutImages int

	for _, prod := range products {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if !s.CheckImageExist(prod.CodeSTU) {
			countWithoutImages++
		}
	}

	stat := &Stat{
		ProductsAll:  len(products),
		WithoutImage: countWithoutImages,
		ImagesAll:    countImages,
	}

	s.statCache = &statCache{
		t:    time.Now(),
		stat: stat,
	}

	return stat, nil
}

func (s *ImagesService) Delete(prodCode int) error {
	return s.repo.Remove(getImageName(prodCode))
}

func getImageName(prodCode int) string {
	return fmt.Sprintf("%d.webp", prodCode)
}
