package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/adriano70/clima-cep/internal/weather"
)

type serviceStub struct {
	result  weather.Temperature
	err     error
	zipcode string
}

func TestManipuladorSerializaTemperaturasSemResiduoDePontoFlutuante(t *testing.T) {
	t.Parallel()

	service := &serviceStub{result: weather.ConvertTemperature(21.2)}
	recorder := request(t, NewHandler(service, discardLogger()), http.MethodGet, "/weather/01001000")

	const want = `{"temp_C":21.2,"temp_F":70.16,"temp_K":294.35}`
	if got := strings.TrimSpace(recorder.Body.String()); got != want {
		t.Fatalf("resposta = %s, esperada %s", got, want)
	}
}

func (s *serviceStub) ByZipcode(_ context.Context, zipcode string) (weather.Temperature, error) {
	s.zipcode = zipcode
	return s.result, s.err
}

func TestManipuladorDeClimaComSucesso(t *testing.T) {
	t.Parallel()

	service := &serviceStub{result: weather.ConvertTemperature(28.5)}
	recorder := request(t, NewHandler(service, discardLogger()), http.MethodGet, "/weather/01001000")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, esperado %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q", got)
	}
	if service.zipcode != "01001000" {
		t.Fatalf("CEP = %q, esperado 01001000", service.zipcode)
	}

	var got weather.Temperature
	if err := json.NewDecoder(recorder.Body).Decode(&got); err != nil {
		t.Fatalf("decodificar resposta: %v", err)
	}
	if got != service.result {
		t.Fatalf("resposta = %#v, esperada %#v", got, service.result)
	}
}

func TestErrosDoManipuladorDeClima(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{name: "CEP inválido", err: weather.ErrInvalidZipcode, wantStatus: 422, wantBody: "invalid zipcode"},
		{name: "CEP não encontrado", err: weather.ErrZipcodeNotFound, wantStatus: 404, wantBody: "can not find zipcode"},
		{name: "erro encapsulado de CEP não encontrado", err: errors.Join(errors.New("provedor"), weather.ErrZipcodeNotFound), wantStatus: 404, wantBody: "can not find zipcode"},
		{name: "serviço externo", err: weather.ErrUpstream, wantStatus: 502, wantBody: "não foi possível consultar o clima"},
	}

	for _, tt := range tests {
		testCase := tt
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			recorder := request(t, NewHandler(&serviceStub{err: testCase.err}, discardLogger()), http.MethodGet, "/weather/99999999")
			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status = %d, esperado %d", recorder.Code, testCase.wantStatus)
			}
			if recorder.Body.String() != testCase.wantBody {
				t.Fatalf("corpo = %q, esperado %q", recorder.Body.String(), testCase.wantBody)
			}
		})
	}
}

func TestMetodoNaoPermitidoNoManipuladorDeClima(t *testing.T) {
	t.Parallel()

	recorder := request(t, NewHandler(&serviceStub{}, discardLogger()), http.MethodPost, "/weather/01001000")
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, esperado %d", recorder.Code, http.StatusMethodNotAllowed)
	}
	if recorder.Header().Get("Allow") != http.MethodGet {
		t.Fatalf("Allow = %q, esperado GET", recorder.Header().Get("Allow"))
	}
}

func TestSaudeECaminhoDesconhecido(t *testing.T) {
	t.Parallel()

	handler := NewHandler(&serviceStub{}, discardLogger())
	for _, path := range []string{"/health", "/healthz"} {
		health := request(t, handler, http.MethodGet, path)
		if health.Code != http.StatusOK || health.Body.String() != "ok" {
			t.Fatalf("resposta de saúde em %s = %d %q, esperada 200 ok", path, health.Code, health.Body.String())
		}
	}

	unknown := request(t, handler, http.MethodGet, "/unknown")
	if unknown.Code != http.StatusNotFound {
		t.Fatalf("status do caminho desconhecido = %d, esperado 404", unknown.Code)
	}
}

func request(t *testing.T, handler http.Handler, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
