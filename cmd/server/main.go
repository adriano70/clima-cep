package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adriano70/clima-cep/internal/config"
	"github.com/adriano70/clima-cep/internal/httpapi"
	"github.com/adriano70/clima-cep/internal/viacep"
	"github.com/adriano70/clima-cep/internal/weather"
	"github.com/adriano70/clima-cep/internal/weatherapi"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, logger); err != nil {
		logger.Error("servidor encerrado", "erro", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("carregar configuração: %w", err)
	}

	httpClient := &http.Client{Timeout: cfg.RequestTimeout}
	locationClient := viacep.NewClient(httpClient, cfg.ViaCEPBaseURL)
	weatherClient := weatherapi.NewClient(httpClient, cfg.WeatherAPIBaseURL, cfg.WeatherAPIKey)
	service := weather.NewService(locationClient, weatherClient)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpapi.NewHandler(service, logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("servidor em execução", "porta", cfg.Port)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("servir HTTP: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("encerrar servidor HTTP: %w", err)
		}
		return nil
	}
}
