package viacep

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/adriano70/clima-cep/internal/weather"
)

const defaultBaseURL = "https://viacep.com.br"

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	httpClient HTTPClient
	baseURL    string
}

func NewClient(httpClient HTTPClient, baseURL string) *Client {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultBaseURL
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    strings.TrimRight(baseURL, "/"),
	}
}

func (c *Client) FindLocation(ctx context.Context, zipcode string) (weather.Location, error) {
	endpoint, err := url.JoinPath(c.baseURL, "ws", zipcode, "json")
	if err != nil {
		return weather.Location{}, fmt.Errorf("montar URL do ViaCEP: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return weather.Location{}, fmt.Errorf("criar requisição do ViaCEP: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return weather.Location{}, fmt.Errorf("%w: requisitar ViaCEP: %v", weather.ErrUpstream, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return weather.Location{}, weather.ErrZipcodeNotFound
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return weather.Location{}, fmt.Errorf("%w: ViaCEP retornou o status %d", weather.ErrUpstream, resp.StatusCode)
	}

	var payload struct {
		City  string       `json:"localidade"`
		State string       `json:"uf"`
		Error flexibleBool `json:"erro"`
	}
	decoder := json.NewDecoder(io.LimitReader(resp.Body, 1<<20))
	if err := decoder.Decode(&payload); err != nil {
		return weather.Location{}, fmt.Errorf("%w: decodificar resposta do ViaCEP: %v", weather.ErrUpstream, err)
	}
	if payload.Error {
		return weather.Location{}, weather.ErrZipcodeNotFound
	}
	if strings.TrimSpace(payload.City) == "" || strings.TrimSpace(payload.State) == "" {
		return weather.Location{}, fmt.Errorf("%w: resposta do ViaCEP sem cidade ou estado", weather.ErrUpstream)
	}

	return weather.Location{City: payload.City, State: payload.State}, nil
}

type flexibleBool bool

func (b *flexibleBool) UnmarshalJSON(data []byte) error {
	switch string(bytes.TrimSpace(data)) {
	case "true", `"true"`:
		*b = true
	case "false", `"false"`:
		*b = false
	default:
		return fmt.Errorf("valor booleano inválido")
	}

	return nil
}
