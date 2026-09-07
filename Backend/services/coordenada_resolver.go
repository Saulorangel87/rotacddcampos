package services

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/empresa/rotas-entrega/models"
)

type coordenadaRuaRepository interface {
	FindAll(ctx context.Context, filters map[string]string) ([]models.Rua, error)
	FindByID(ctx context.Context, id uint) (*models.Rua, error)
}

type geocodificacaoCache interface {
	FindByChave(ctx context.Context, chave string) (*models.GeocodificacaoRua, error)
	Save(ctx context.Context, geocodificacao *models.GeocodificacaoRua) error
}

type Geocodificador interface {
	BuscarRua(ctx context.Context, nomeRua string) (*CoordenadaReferencia, error)
}

type SolicitacaoCoordenada struct {
	ChaveAgrupamento string
	NomeRua          string
	RuaID            *uint
	PermitirExterno  bool
}

type CoordenadaReferencia struct {
	Latitude  float64
	Longitude float64
	Fonte     string
	// Pontos são os vértices do traçado interno. A coordenada média continua
	// disponível para exibição e compatibilidade, enquanto o otimizador usa os
	// pontos para medir a menor distância até a rua.
	Pontos []PontoGeografico
}

type CoordenadaResolver interface {
	Resolver(ctx context.Context, solicitacao SolicitacaoCoordenada) (*CoordenadaReferencia, error)
}

type coordenadaResolver struct {
	ruas           coordenadaRuaRepository
	cache          geocodificacaoCache
	geocodificador Geocodificador
}

const validadeCacheNaoEncontrada = 7 * 24 * time.Hour

func NewCoordenadaResolver(ruas coordenadaRuaRepository, cache geocodificacaoCache, geocodificador Geocodificador) CoordenadaResolver {
	return &coordenadaResolver{ruas: ruas, cache: cache, geocodificador: geocodificador}
}

func (r *coordenadaResolver) Resolver(ctx context.Context, solicitacao SolicitacaoCoordenada) (*CoordenadaReferencia, error) {
	if solicitacao.ChaveAgrupamento == "" || solicitacao.NomeRua == "" {
		return nil, nil
	}

	ruas, err := r.ruasDaSolicitacao(ctx, solicitacao)
	if err != nil {
		return nil, err
	}
	if coordenada := coordenadaDasGeometrias(ruas); coordenada != nil {
		return coordenada, nil
	}

	emCache, err := r.cache.FindByChave(ctx, solicitacao.ChaveAgrupamento)
	if err != nil {
		return nil, err
	}
	if emCache != nil {
		if emCache.Encontrada {
			return &CoordenadaReferencia{
				Latitude: emCache.Latitude, Longitude: emCache.Longitude, Fonte: emCache.Fonte,
			}, nil
		}
		if time.Since(emCache.UpdatedAt) < validadeCacheNaoEncontrada {
			return nil, nil
		}
	}
	if !solicitacao.PermitirExterno || r.geocodificador == nil {
		return nil, nil
	}

	coordenada, err := r.geocodificador.BuscarRua(ctx, solicitacao.NomeRua)
	if err != nil {
		return nil, err
	}
	if coordenada == nil {
		err = r.cache.Save(ctx, &models.GeocodificacaoRua{
			ChaveAgrupamento: solicitacao.ChaveAgrupamento,
			NomeRua:          solicitacao.NomeRua,
			Fonte:            models.FonteCoordenadaNominatim,
			Encontrada:       false,
		})
		return nil, err
	}
	if err := r.cache.Save(ctx, &models.GeocodificacaoRua{
		ChaveAgrupamento: solicitacao.ChaveAgrupamento,
		NomeRua:          solicitacao.NomeRua,
		Latitude:         coordenada.Latitude,
		Longitude:        coordenada.Longitude,
		Fonte:            coordenada.Fonte,
		Encontrada:       true,
	}); err != nil {
		return nil, err
	}
	return coordenada, nil
}

func (r *coordenadaResolver) ruasDaSolicitacao(ctx context.Context, solicitacao SolicitacaoCoordenada) ([]models.Rua, error) {
	if strings.HasPrefix(solicitacao.ChaveAgrupamento, "rua:") && solicitacao.RuaID != nil {
		rua, err := r.ruas.FindByID(ctx, *solicitacao.RuaID)
		if err != nil {
			return nil, err
		}
		if rua == nil {
			return nil, nil
		}
		return []models.Rua{*rua}, nil
	}

	base := normalizarNomeBase(solicitacao.NomeRua)
	candidatas, err := r.ruas.FindAll(ctx, map[string]string{"nome": melhorTokenBusca(base)})
	if err != nil {
		return nil, err
	}
	ruasEncontradas := make([]models.Rua, 0, len(candidatas))
	for _, rua := range candidatas {
		if normalizarNomeBase(rua.NomeRua) == base {
			ruasEncontradas = append(ruasEncontradas, rua)
		}
	}
	return ruasEncontradas, nil
}

func coordenadaDasGeometrias(ruas []models.Rua) *CoordenadaReferencia {
	var somaLatitude, somaLongitude float64
	var quantidade int
	pontosGeograficos := make([]PontoGeografico, 0)
	for _, rua := range ruas {
		pontos, err := extrairPontosGeoJSON(rua.Geometria)
		if err != nil {
			continue
		}
		for _, ponto := range pontos {
			somaLongitude += ponto[0]
			somaLatitude += ponto[1]
			quantidade++
			pontosGeograficos = append(pontosGeograficos, PontoGeografico{Latitude: ponto[1], Longitude: ponto[0]})
		}
	}
	if quantidade == 0 {
		return nil
	}
	return &CoordenadaReferencia{
		Latitude:  somaLatitude / float64(quantidade),
		Longitude: somaLongitude / float64(quantidade),
		Fonte:     models.FonteCoordenadaGeometria,
		Pontos:    pontosGeograficos,
	}
}

func extrairPontosGeoJSON(valor string) ([][2]float64, error) {
	if strings.TrimSpace(valor) == "" {
		return nil, errors.New("geometria vazia")
	}
	var geometria struct {
		Type        string          `json:"type"`
		Coordinates json.RawMessage `json:"coordinates"`
	}
	if err := json.Unmarshal([]byte(valor), &geometria); err != nil {
		return nil, err
	}

	var pontos [][2]float64
	switch geometria.Type {
	case "Point":
		var ponto []float64
		if err := json.Unmarshal(geometria.Coordinates, &ponto); err != nil {
			return nil, err
		}
		adicionarPonto(&pontos, ponto)
	case "LineString":
		var linha [][]float64
		if err := json.Unmarshal(geometria.Coordinates, &linha); err != nil {
			return nil, err
		}
		for _, ponto := range linha {
			adicionarPonto(&pontos, ponto)
		}
	case "MultiLineString":
		var linhas [][][]float64
		if err := json.Unmarshal(geometria.Coordinates, &linhas); err != nil {
			return nil, err
		}
		for _, linha := range linhas {
			for _, ponto := range linha {
				adicionarPonto(&pontos, ponto)
			}
		}
	default:
		return nil, errors.New("tipo de geometria não suportado")
	}
	if len(pontos) == 0 {
		return nil, errors.New("geometria sem pontos válidos")
	}
	return pontos, nil
}

func adicionarPonto(destino *[][2]float64, ponto []float64) {
	if len(ponto) < 2 || math.IsNaN(ponto[0]) || math.IsNaN(ponto[1]) || math.IsInf(ponto[0], 0) || math.IsInf(ponto[1], 0) {
		return
	}
	if ponto[0] < -180 || ponto[0] > 180 || ponto[1] < -90 || ponto[1] > 90 {
		return
	}
	*destino = append(*destino, [2]float64{ponto[0], ponto[1]})
}
