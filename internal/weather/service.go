package weather

import (
	"context"
	"errors"
	"fmt"
	"math"
)

var (
	ErrInvalidZipcode  = errors.New("invalid zipcode")
	ErrZipcodeNotFound = errors.New("can not find zipcode")
	ErrUpstream        = errors.New("erro no serviço externo")
)

type Location struct {
	City  string
	State string
}

type Temperature struct {
	Celsius    float64 `json:"temp_C"`
	Fahrenheit float64 `json:"temp_F"`
	Kelvin     float64 `json:"temp_K"`
}

type LocationFinder interface {
	FindLocation(ctx context.Context, zipcode string) (Location, error)
}

type WeatherFinder interface {
	CurrentCelsius(ctx context.Context, location Location) (float64, error)
}

type Service struct {
	locations LocationFinder
	weather   WeatherFinder
}

func NewService(locations LocationFinder, weather WeatherFinder) *Service {
	return &Service{locations: locations, weather: weather}
}

func (s *Service) ByZipcode(ctx context.Context, zipcode string) (Temperature, error) {
	if !ValidZipcode(zipcode) {
		return Temperature{}, ErrInvalidZipcode
	}

	location, err := s.locations.FindLocation(ctx, zipcode)
	if err != nil {
		return Temperature{}, fmt.Errorf("localizar cidade: %w", err)
	}

	celsius, err := s.weather.CurrentCelsius(ctx, location)
	if err != nil {
		return Temperature{}, fmt.Errorf("consultar clima atual: %w", err)
	}

	return ConvertTemperature(celsius), nil
}

func ValidZipcode(zipcode string) bool {
	if len(zipcode) != 8 {
		return false
	}

	for i := 0; i < len(zipcode); i++ {
		if zipcode[i] < '0' || zipcode[i] > '9' {
			return false
		}
	}

	return true
}

func ConvertTemperature(celsius float64) Temperature {
	return Temperature{
		Celsius:    celsius,
		Fahrenheit: roundToTwoDecimalPlaces(celsius*1.8 + 32),
		Kelvin:     roundToTwoDecimalPlaces(celsius + 273.15),
	}
}

func roundToTwoDecimalPlaces(value float64) float64 {
	return math.Round(value*100) / 100
}
