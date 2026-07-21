package weatherapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adriano70/clima-cep/internal/weather"
)

func TestTemperaturaAtualEmCelsius(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/current.json" {
			t.Errorf("caminho = %q, esperado /current.json", r.URL.Path)
		}
		if got := r.URL.Query().Get("key"); got != "test-key" {
			t.Errorf("chave = %q, esperada test-key", got)
		}
		if got := r.URL.Query().Get("q"); got != "São Paulo, SP, Brazil" {
			t.Errorf("q = %q, esperado São Paulo, SP, Brazil", got)
		}
		if got := r.URL.Query().Get("aqi"); got != "no" {
			t.Errorf("aqi = %q, esperado no", got)
		}
		_, _ = w.Write([]byte(`{"current":{"temp_c":28.5}}`))
	}))
	defer server.Close()

	got, err := NewClient(server.Client(), server.URL, "test-key").CurrentCelsius(
		context.Background(),
		weather.Location{City: "São Paulo", State: "SP"},
	)
	if err != nil {
		t.Fatalf("CurrentCelsius() retornou erro = %v", err)
	}
	if got != 28.5 {
		t.Fatalf("CurrentCelsius() = %v, esperado 28.5", got)
	}
}

func TestErrosDoServicoExternoAoConsultarTemperatura(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status int
		body   string
	}{
		{name: "status do servidor", status: http.StatusUnauthorized, body: `{}`},
		{name: "JSON malformado", status: http.StatusOK, body: `{`},
		{name: "temperatura ausente", status: http.StatusOK, body: `{"current":{}}`},
	}

	for _, tt := range tests {
		testCase := tt
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(testCase.status)
				_, _ = w.Write([]byte(testCase.body))
			}))
			defer server.Close()

			_, err := NewClient(server.Client(), server.URL, "test-key").CurrentCelsius(
				context.Background(),
				weather.Location{City: "São Paulo", State: "SP"},
			)
			if !errors.Is(err, weather.ErrUpstream) {
				t.Fatalf("CurrentCelsius() retornou erro = %v, esperado ErrUpstream", err)
			}
		})
	}
}
