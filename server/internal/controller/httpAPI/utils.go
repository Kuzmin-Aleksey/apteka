package httpAPI

import (
	"encoding/json"
	"html/template"
	"net/http"
	"server/pkg/errcodes"
	"server/pkg/failure"
)

func (h *Handler) writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.l.Println("json encoding error:", err.Error())
	}
}

type responseError struct {
	Error string `json:"error"`
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	h.l.Println(err)
	w.Header().Set("Content-Type", "application/json")

	code, status := getErrorStatus(err)

	w.WriteHeader(status)

	if e := json.NewEncoder(w).Encode(responseError{code.String()}); e != nil {
		h.l.Println("json encoding error:", e)
	}
}

func (h *Handler) handleTemplate(tmpPath ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmp, err := template.ParseFiles(tmpPath...)
		if err != nil {
			h.writeError(w, failure.NewInternalError("parse template: "+err.Error()))
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmp.Execute(w, nil); err != nil {
			h.writeError(w, failure.NewInternalError("execute template: "+err.Error()))
			return
		}
	}
}

func getErrorStatus(err error) (errcodes.Code, int) {
	switch {
	case failure.IsNotFoundError(err):
		return errcodes.ErrNotFound, http.StatusNotFound
	case failure.IsInvalidRequestError(err):
		return errcodes.ErrInvalidRequest, http.StatusBadRequest
	case failure.IsInvalidFileError(err):
		return errcodes.ErrInvalidFile, http.StatusBadRequest
	case failure.IsUnauthorizedError(err):
		return errcodes.ErrUnauthorized, http.StatusUnauthorized
	case failure.IsLockedError(err):
		return errcodes.ErrLocked, http.StatusLocked
	default:
		return errcodes.ErrUnknown, http.StatusInternalServerError
	}
}
