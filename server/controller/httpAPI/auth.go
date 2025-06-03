package httpAPI

import (
	"net/http"
	"server/pkg/failure"
	"strings"
)

type LoginInRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) ApiHandleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.writeError(w, failure.NewInvalidRequestError("invalid form"+": "+err.Error()))
		return
	}

	session, err := h.auth.Login(r.Context(), r.Form.Get("username"), r.Form.Get("password"))
	if err != nil {
		h.writeError(w, err)
		return
	}

	w.Header().Add("Token", session)

	// if user logged, try logout
	typeAndToken := strings.Split(r.Header.Get("Authorization"), " ")
	if len(typeAndToken) != 2 {
		return
	}
	if err := h.auth.Logout(r.Context(), typeAndToken[1]); err != nil {
		h.l.Println(err)
	}
}

func (h *Handler) ApiHandleLogout(w http.ResponseWriter, r *http.Request) {
	typeAndToken := strings.Split(r.Header.Get("Authorization"), " ")
	if len(typeAndToken) != 2 {
		return
	}

	if err := h.auth.Logout(r.Context(), typeAndToken[1]); err != nil {
		h.writeError(w, err)
	}
}
