package weather

import (
	"context"
	"errors"
	"math"
	"testing"
)

type locationFinderStub struct {
	location Location
	err      error
	calls    int
}

func (s *locationFinderStub) FindLocation(context.Context, string) (Location, error) {
	s.calls++
	return s.location, s.err
}

type weatherFinderStub struct {
	celsius  float64
	err      error
	location Location
	calls    int
}

func (s *weatherFinderStub) CurrentCelsius(_ context.Context, location Location) (float64, error) {
	s.calls++
	s.location = location
	return s.celsius, s.err
}

func TestValidarCEP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		zipcode string
		want    bool
	}{
		{name: "oito dígitos", zipcode: "01001000", want: true},
		{name: "muito curto", zipcode: "0100100", want: false},
		{name: "muito longo", zipcode: "010010000", want: false},
		{name: "letra", zipcode: "01001A00", want: false},
		{name: "hífen", zipcode: "01001-000", want: false},
		{name: "dígito Unicode", zipcode: "０1001000", want: false},
		{name: "vazio", zipcode: "", want: false},
	}

	for _, tt := range tests {
		testCase := tt
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := ValidZipcode(testCase.zipcode); got != testCase.want {
				t.Fatalf("ValidZipcode(%q) = %v, esperado %v", testCase.zipcode, got, testCase.want)
			}
		})
	}
}

func TestConverterTemperatura(t *testing.T) {
	t.Parallel()

	got := ConvertTemperature(28.5)
	assertFloat(t, got.Celsius, 28.5)
	assertFloat(t, got.Fahrenheit, 83.3)
	assertFloat(t, got.Kelvin, 301.65)
}

func TestServicoPorCEP(t *testing.T) {
	t.Parallel()

	locations := &locationFinderStub{location: Location{City: "São Paulo", State: "SP"}}
	weatherFinder := &weatherFinderStub{celsius: 28.5}
	service := NewService(locations, weatherFinder)

	got, err := service.ByZipcode(context.Background(), "01001000")
	if err != nil {
		t.Fatalf("ByZipcode() retornou erro = %v", err)
	}
	if locations.calls != 1 || weatherFinder.calls != 1 {
		t.Fatalf("chamadas aos provedores = localização %d, clima %d; esperado 1 para cada", locations.calls, weatherFinder.calls)
	}
	if weatherFinder.location != locations.location {
		t.Fatalf("localização meteorológica = %#v, esperada %#v", weatherFinder.location, locations.location)
	}
	assertFloat(t, got.Fahrenheit, 83.3)
}

func TestServicoRejeitaCEPInvalidoAntesDosProvedores(t *testing.T) {
	t.Parallel()

	locations := &locationFinderStub{}
	weatherFinder := &weatherFinderStub{}
	service := NewService(locations, weatherFinder)

	_, err := service.ByZipcode(context.Background(), "01001-000")
	if !errors.Is(err, ErrInvalidZipcode) {
		t.Fatalf("ByZipcode() retornou erro = %v, esperado ErrInvalidZipcode", err)
	}
	if locations.calls != 0 || weatherFinder.calls != 0 {
		t.Fatal("provedores foram chamados para um CEP inválido")
	}
}

func TestServicoPreservaIdentidadeDoErroDoProvedor(t *testing.T) {
	t.Parallel()

	service := NewService(&locationFinderStub{err: ErrZipcodeNotFound}, &weatherFinderStub{})
	_, err := service.ByZipcode(context.Background(), "99999999")
	if !errors.Is(err, ErrZipcodeNotFound) {
		t.Fatalf("ByZipcode() retornou erro = %v, esperado ErrZipcodeNotFound", err)
	}
}

func assertFloat(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("obtido %.12f, esperado %.12f", got, want)
	}
}
