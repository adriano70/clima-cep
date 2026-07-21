package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPort           = "8080"
	defaultRequestTimeout = 5 * time.Second
	defaultViaCEPBaseURL  = "https://viacep.com.br"
	defaultWeatherBaseURL = "https://api.weatherapi.com/v1"
)

type Config struct {
	Port              string
	WeatherAPIKey     string
	RequestTimeout    time.Duration
	ViaCEPBaseURL     string
	WeatherAPIBaseURL string
}

func Load() (Config, error) {
	config := Config{
		Port:              valueOrDefault("PORT", defaultPort),
		WeatherAPIKey:     strings.TrimSpace(os.Getenv("WEATHER_API_KEY")),
		RequestTimeout:    defaultRequestTimeout,
		ViaCEPBaseURL:     valueOrDefault("VIACEP_BASE_URL", defaultViaCEPBaseURL),
		WeatherAPIBaseURL: valueOrDefault("WEATHER_API_BASE_URL", defaultWeatherBaseURL),
	}

	if config.WeatherAPIKey == "" {
		return Config{}, fmt.Errorf("WEATHER_API_KEY é obrigatória")
	}

	port, err := strconv.Atoi(config.Port)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("PORT deve ser um número entre 1 e 65535")
	}

	if rawTimeout := strings.TrimSpace(os.Getenv("HTTP_TIMEOUT")); rawTimeout != "" {
		config.RequestTimeout, err = time.ParseDuration(rawTimeout)
		if err != nil || config.RequestTimeout <= 0 {
			return Config{}, fmt.Errorf("HTTP_TIMEOUT deve ser uma duração positiva")
		}
	}

	return config, nil
}

func valueOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
