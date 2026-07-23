package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/adriano70/clima-cep/internal/weather"
)

type WeatherService interface {
	ByZipcode(ctx context.Context, zipcode string) (weather.Temperature, error)
}

type Handler struct {
	service WeatherService
	logger  *slog.Logger
}

func NewHandler(service WeatherService, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	h := &Handler{service: service, logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.health)
	mux.HandleFunc("/healthz", h.health)
	mux.HandleFunc("/weather/", h.weather)
	return mux
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeText(w, http.StatusMethodNotAllowed, "método não permitido")
		return
	}

	writeText(w, http.StatusOK, "ok")
}

func (h *Handler) weather(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeText(w, http.StatusMethodNotAllowed, "método não permitido")
		return
	}

	zipcode := strings.TrimPrefix(r.URL.Path, "/weather/")
	result, err := h.service.ByZipcode(r.Context(), zipcode)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		h.logger.Error("codificar resposta meteorológica", "erro", err)
	}
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, weather.ErrInvalidZipcode):
		writeText(w, http.StatusUnprocessableEntity, weather.ErrInvalidZipcode.Error())
	case errors.Is(err, weather.ErrZipcodeNotFound):
		writeText(w, http.StatusNotFound, weather.ErrZipcodeNotFound.Error())
	default:
		h.logger.Error("falha na consulta meteorológica", "erro", err)
		writeText(w, http.StatusBadGateway, "não foi possível consultar o clima")
	}
}

func writeText(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, body)
}
