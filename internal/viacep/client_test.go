package viacep

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adriano70/clima-cep/internal/weather"
)

func TestLocalizarCidade(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ws/01001000/json" {
			t.Errorf("caminho = %q, esperado /ws/01001000/json", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"localidade":"São Paulo","uf":"SP"}`))
	}))
	defer server.Close()

	got, err := NewClient(server.Client(), server.URL).FindLocation(context.Background(), "01001000")
	if err != nil {
		t.Fatalf("FindLocation() retornou erro = %v", err)
	}
	want := weather.Location{City: "São Paulo", State: "SP"}
	if got != want {
		t.Fatalf("FindLocation() = %#v, esperado %#v", got, want)
	}
}

func TestCidadeNaoEncontrada(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"erro":true}`))
	}))
	defer server.Close()

	_, err := NewClient(server.Client(), server.URL).FindLocation(context.Background(), "99999999")
	if !errors.Is(err, weather.ErrZipcodeNotFound) {
		t.Fatalf("FindLocation() retornou erro = %v, esperado ErrZipcodeNotFound", err)
	}
}

func TestErrosDoServicoExternoAoLocalizarCidade(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status int
		body   string
	}{
		{name: "status do servidor", status: http.StatusServiceUnavailable, body: `{}`},
		{name: "JSON malformado", status: http.StatusOK, body: `{`},
		{name: "cidade ausente", status: http.StatusOK, body: `{"uf":"SP"}`},
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

			_, err := NewClient(server.Client(), server.URL).FindLocation(context.Background(), "01001000")
			if !errors.Is(err, weather.ErrUpstream) {
				t.Fatalf("FindLocation() retornou erro = %v, esperado ErrUpstream", err)
			}
		})
	}
}
