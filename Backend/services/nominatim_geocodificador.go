package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/empresa/rotas-entrega/models"
)

const (
	municipioGeocodificacao = "Campos dos Goytacazes"
	estadoGeocodificacao    = "Rio de Janeiro"
	paisGeocodificacao      = "Brasil"
)

type nominatimGeocodificador struct {
	baseURL     string
	userAgent   string
	httpClient  *http.Client
	mu          sync.Mutex
	ultimaBusca time.Time
}

func NewNominatimGeocodificador(baseURL, userAgent string) Geocodificador {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" || strings.EqualFold(baseURL, "off") {
		return nil
	}
	return &nominatimGeocodificador{
		baseURL:    strings.TrimRight(baseURL, "/"),
		userAgent:  userAgent,
		httpClient: &http.Client{Timeout: 8 * time.Second},
	}
}

func (g *nominatimGeocodificador) BuscarRua(ctx context.Context, nomeRua string) (*CoordenadaReferencia, error) {
	if err := g.aguardarLimite(ctx); err != nil {
		return nil, err
	}

	parametros := url.Values{}
	parametros.Set("format", "jsonv2")
	parametros.Set("limit", "5")
	parametros.Set("addressdetails", "1")
	parametros.Set("countrycodes", "br")
	parametros.Set("street", nomeRua)
	parametros.Set("city", municipioGeocodificacao)
	parametros.Set("state", estadoGeocodificacao)
	parametros.Set("country", paisGeocodificacao)

	requisicao, err := http.NewRequestWithContext(ctx, http.MethodGet, g.baseURL+"/search?"+parametros.Encode(), nil)
	if err != nil {
		return nil, err
	}
	requisicao.Header.Set("User-Agent", g.userAgent)
	requisicao.Header.Set("Accept", "application/json")

	resposta, err := g.httpClient.Do(requisicao)
	if err != nil {
		return nil, err
	}
	defer resposta.Body.Close()
	if resposta.StatusCode < 200 || resposta.StatusCode >= 300 {
		return nil, fmt.Errorf("geocodificador respondeu HTTP %d", resposta.StatusCode)
	}

	var resultados []struct {
		Lat         string            `json:"lat"`
		Lon         string            `json:"lon"`
		DisplayName string            `json:"display_name"`
		Address     map[string]string `json:"address"`
	}
	if err := json.NewDecoder(resposta.Body).Decode(&resultados); err != nil {
		return nil, err
	}
	for _, resultado := range resultados {
		latitude, erroLatitude := strconv.ParseFloat(resultado.Lat, 64)
		longitude, erroLongitude := strconv.ParseFloat(resultado.Lon, 64)
		if erroLatitude != nil || erroLongitude != nil || !resultadoPertenceACampos(resultado.DisplayName, resultado.Address, latitude, longitude) {
			continue
		}
		return &CoordenadaReferencia{Latitude: latitude, Longitude: longitude, Fonte: models.FonteCoordenadaNominatim}, nil
	}
	return nil, nil
}

func (g *nominatimGeocodificador) aguardarLimite(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if espera := time.Second - time.Since(g.ultimaBusca); espera > 0 {
		timer := time.NewTimer(espera)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
	}
	g.ultimaBusca = time.Now()
	return nil
}

func resultadoPertenceACampos(displayName string, endereco map[string]string, latitude, longitude float64) bool {
	// Limites amplos do município impedem aceitar homônimos de outros estados.
	if latitude < -22.2 || latitude > -21.1 || longitude < -42.0 || longitude > -40.7 {
		return false
	}
	municipio := normalizarTexto(municipioGeocodificacao)
	for _, campo := range []string{"city", "town", "municipality", "county", "state_district"} {
		if strings.Contains(normalizarTexto(endereco[campo]), municipio) {
			return true
		}
	}
	return strings.Contains(normalizarTexto(displayName), municipio)
}
