package weatherapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/adriano70/clima-cep/internal/weather"
)

const defaultBaseURL = "https://api.weatherapi.com/v1"

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	httpClient HTTPClient
	baseURL    string
	apiKey     string
}

func NewClient(httpClient HTTPClient, baseURL, apiKey string) *Client {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultBaseURL
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
	}
}

func (c *Client) CurrentCelsius(ctx context.Context, location weather.Location) (float64, error) {
	endpoint, err := url.JoinPath(c.baseURL, "current.json")
	if err != nil {
		return 0, fmt.Errorf("montar URL da WeatherAPI: %w", err)
	}

	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return 0, fmt.Errorf("interpretar URL da WeatherAPI: %w", err)
	}
	query := parsedURL.Query()
	query.Set("key", c.apiKey)
	query.Set("q", strings.Join([]string{location.City, location.State, "Brazil"}, ", "))
	query.Set("aqi", "no")
	parsedURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return 0, fmt.Errorf("criar requisição da WeatherAPI: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("%w: requisitar WeatherAPI: %v", weather.ErrUpstream, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return 0, fmt.Errorf("%w: WeatherAPI retornou o status %d", weather.ErrUpstream, resp.StatusCode)
	}

	var payload struct {
		Current struct {
			TempC *float64 `json:"temp_c"`
		} `json:"current"`
	}
	decoder := json.NewDecoder(io.LimitReader(resp.Body, 1<<20))
	if err := decoder.Decode(&payload); err != nil {
		return 0, fmt.Errorf("%w: decodificar resposta da WeatherAPI: %v", weather.ErrUpstream, err)
	}
	if payload.Current.TempC == nil {
		return 0, fmt.Errorf("%w: resposta da WeatherAPI sem temperatura", weather.ErrUpstream)
	}

	return *payload.Current.TempC, nil
}
