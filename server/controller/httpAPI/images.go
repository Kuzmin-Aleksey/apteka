package httpAPI

import (
	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
	"net/http"
	"server/internal/infrastructure/integration/image_parser"
	"server/pkg/failure"
	"strconv"
	"strings"
	"time"
)

var wsUpgrader = &websocket.Upgrader{
	ReadBufferSize:  256,
	WriteBufferSize: 256,
}

func (h *Handler) ApiLoadImages(w http.ResponseWriter, r *http.Request) {
	if err := h.images.LoadImages(context.Background()); err != nil {
		h.writeError(w, err)
		return
	}
}

func (h *Handler) ApiLoadImagesProgress(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		h.l.Println("failed to upgrade websocket connection: ", err)
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
				h.l.Println("write websocket: ", err)
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

func (h *Handler) ApiImagesStopLoading(w http.ResponseWriter, r *http.Request) {
	if err := h.images.StopLoading(r.Context()); err != nil {
		h.writeError(w, err)
		return
	}
}

func (h *Handler) ApiSaveImage(w http.ResponseWriter, r *http.Request) {
	imgType := r.Header.Get("Content-Type")

	if strings.HasPrefix("image/", imgType) {
		h.writeError(w, failure.NewInvalidRequestError("invalid image type: "+imgType))
		return
	}

	prodCode, err := strconv.Atoi(r.FormValue("product_code"))
	if err != nil {
		h.writeError(w, failure.NewInvalidRequestError("invalid product code: "+r.FormValue("product_code")+": "+err.Error()))
		return
	}

	img, err := image_parser.ConvertToWebp(r.Body, imgType)
	if err != nil {
		h.writeError(w, failure.NewInternalError("cannot convert image: "+err.Error()))
		return
	}

	if err := h.images.SaveWebpImage(prodCode, img); err != nil {
		h.writeError(w, err)
		return
	}
}

func (h *Handler) ApiGetImagesStat(w http.ResponseWriter, r *http.Request) {
	stat, err := h.images.GetStat(r.Context())
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, stat)
}

func (h *Handler) ApiCheckImage(w http.ResponseWriter, r *http.Request) {
	prodCode, err := strconv.Atoi(r.FormValue("product_code"))
	if err != nil {
		h.writeError(w, failure.NewInvalidRequestError("invalid product code: "+r.FormValue("product_code")+": "+err.Error()))
		return
	}

	ok := h.images.CheckImageExist(prodCode)

	h.writeJSON(w, ok)
}

func (h *Handler) ApiDeleteImage(w http.ResponseWriter, r *http.Request) {
	prodCode, err := strconv.Atoi(r.FormValue("product_code"))
	if err != nil {
		h.writeError(w, failure.NewInvalidRequestError("invalid product code: "+r.FormValue("product_code")+": "+err.Error()))
		return
	}

	if err := h.images.Delete(prodCode); err != nil {
		h.writeError(w, err)
		return
	}
}
