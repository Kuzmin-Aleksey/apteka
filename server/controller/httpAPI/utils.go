package httpAPI

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"server/domain/models"
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

	domainErr := models.GetDomainErr(err)
	w.WriteHeader(getErrorStatus(domainErr))

	if e := json.NewEncoder(w).Encode(responseError{domainErr.Error()}); e != nil {
		h.l.Println("json encoding error:", e)
	}
}

func (h *Handler) handleTemplate(tmpPath ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmp, err := template.ParseFiles(tmpPath...)
		if err != nil {
			h.writeError(w, models.NewError(models.ErrUnknown, "parse template", err))
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmp.Execute(w, nil); err != nil {
			h.writeError(w, models.NewError(models.ErrUnknown, "execute template", err))
			return
		}
	}
}

func getErrorStatus(err error) int {
	switch {
	case errors.Is(err, models.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, models.ErrInvalidRequest), errors.Is(err, models.ErrInvalidFile):
		return http.StatusBadRequest
	case errors.Is(err, models.ErrInvalidLoginOrPassword), errors.Is(err, models.ErrUnauthorized):
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
