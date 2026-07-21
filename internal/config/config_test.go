package config

import (
	"strings"
	"testing"
	"time"
)

func TestCarregar(t *testing.T) {
	t.Setenv("WEATHER_API_KEY", "segredo")
	t.Setenv("PORT", "9090")
	t.Setenv("HTTP_TIMEOUT", "2s")
	t.Setenv("VIACEP_BASE_URL", "http://viacep.test")
	t.Setenv("WEATHER_API_BASE_URL", "http://weather.test")

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() retornou erro = %v", err)
	}
	if got.Port != "9090" || got.WeatherAPIKey != "segredo" || got.RequestTimeout != 2*time.Second {
		t.Fatalf("Load() = %#v", got)
	}
	if got.ViaCEPBaseURL != "http://viacep.test" || got.WeatherAPIBaseURL != "http://weather.test" {
		t.Fatalf("URLs base = %q e %q", got.ViaCEPBaseURL, got.WeatherAPIBaseURL)
	}
}

func TestCarregarExigeChaveDaWeatherAPI(t *testing.T) {
	t.Setenv("WEATHER_API_KEY", "")
	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "WEATHER_API_KEY") {
		t.Fatalf("Load() retornou erro = %v, esperada ausência de WEATHER_API_KEY", err)
	}
}

func TestCarregarRejeitaValoresInvalidos(t *testing.T) {
	tests := []struct {
		name    string
		port    string
		timeout string
	}{
		{name: "porta inválida", port: "nao-e-uma-porta"},
		{name: "porta fora do intervalo", port: "70000"},
		{name: "limite de tempo inválido", port: "8080", timeout: "algum-dia"},
		{name: "limite de tempo negativo", port: "8080", timeout: "-1s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("WEATHER_API_KEY", "segredo")
			t.Setenv("PORT", tt.port)
			t.Setenv("HTTP_TIMEOUT", tt.timeout)
			if _, err := Load(); err == nil {
				t.Fatal("Load() não retornou erro; era esperado um erro de validação")
			}
		})
	}
}
