package server

import (
	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
	"net/http"
	"server/internal/domain/service/images"
	"server/internal/infrastructure/integration/image_parser"
	"server/pkg/contextx"
	"server/pkg/failure"
	"server/pkg/logx"
	"strconv"
	"strings"
	"time"
)

type ImagesServer struct {
	images *images.ImagesService
}

func NewImagesServer(images *images.ImagesService) *ImagesServer {
	return &ImagesServer{images: images}
}

var wsUpgrader = &websocket.Upgrader{
	ReadBufferSize:  256,
	WriteBufferSize: 256,
}

func (h *ImagesServer) ApiLoadImages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = contextx.WithLogger(context.Background(), contextx.GetLoggerOrDefault(ctx))

	if err := h.images.LoadImages(ctx); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
}

func (h *ImagesServer) ApiLoadImagesProgress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := contextx.GetLoggerOrDefault(ctx)

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		l.Error("upgrade websocket", logx.Error(err))
		return
	}

	stopCh := make(chan struct{})

	go func() {
		t, mess, err := conn.ReadMessage()
		if err != nil || string(mess) == "close" || t != websocket.CloseNormalClosure {
			close(stopCh)
		}
	}()

	go func() {
		defer conn.Close()

		ticker := time.NewTicker(time.Second)
		for h.images.IsLoading {
			if err := conn.WriteJSON(h.images.CurrentProgress); err != nil {
				l.Error("write websocket", logx.Error(err))
				return
			}
			select {
			case <-ticker.C:
			case <-stopCh:
				return
			}
		}
		conn.WriteJSON(-1)
	}()
}

func (h *ImagesServer) ApiImagesStopLoading(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := h.images.StopLoading(ctx); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
}

func (h *ImagesServer) ApiSaveImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	imgType := r.Header.Get("Content-Type")

	if strings.HasPrefix("image/", imgType) {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid image type: "+imgType))
		return
	}

	prodCode, err := strconv.Atoi(r.FormValue("product_code"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid product code: "+r.FormValue("product_code")+": "+err.Error()))
		return
	}

	img, err := image_parser.ConvertToWebp(r.Body, imgType)
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInternalError("cannot convert image: "+err.Error()))
		return
	}

	if err := h.images.SaveWebpImage(prodCode, img); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
}

func (h *ImagesServer) ApiGetImagesStat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stat, err := h.images.GetStat(ctx)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}

	writeJson(ctx, w, stat, http.StatusOK)
}

func (h *ImagesServer) ApiCheckImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	prodCode, err := strconv.Atoi(r.FormValue("product_code"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid product code: "+r.FormValue("product_code")+": "+err.Error()))
		return
	}

	ok := h.images.CheckImageExist(prodCode)

	writeJson(ctx, w, ok, http.StatusOK)
}

func (h *ImagesServer) ApiDeleteImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	prodCode, err := strconv.Atoi(r.FormValue("product_code"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid product code: "+r.FormValue("product_code")+": "+err.Error()))
		return
	}

	if err := h.images.Delete(prodCode); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
}
