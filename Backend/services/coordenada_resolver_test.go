package services

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/empresa/rotas-entrega/models"
)

type coordenadaRuaRepoFake struct {
	ruas []models.Rua
}

func (r coordenadaRuaRepoFake) FindAll(_ context.Context, _ map[string]string) ([]models.Rua, error) {
	return r.ruas, nil
}

func (r coordenadaRuaRepoFake) FindByID(_ context.Context, id uint) (*models.Rua, error) {
	for _, rua := range r.ruas {
		if rua.ID == id {
			copia := rua
			return &copia, nil
		}
	}
	return nil, nil
}

type geocodificacaoCacheFake struct {
	item       *models.GeocodificacaoRua
	quantidade int
}

func (c *geocodificacaoCacheFake) FindByChave(_ context.Context, _ string) (*models.GeocodificacaoRua, error) {
	return c.item, nil
}

func (c *geocodificacaoCacheFake) Save(_ context.Context, item *models.GeocodificacaoRua) error {
	c.item = item
	c.item.UpdatedAt = time.Now()
	c.quantidade++
	return nil
}

type geocodificadorFake struct {
	coordenada *CoordenadaReferencia
	chamadas   int
}

func (g *geocodificadorFake) BuscarRua(_ context.Context, _ string) (*CoordenadaReferencia, error) {
	g.chamadas++
	return g.coordenada, nil
}

func TestCoordenadaResolverUsaGeometriaInternaAntesDoFallback(t *testing.T) {
	ruaID := uint(10)
	ruAs := coordenadaRuaRepoFake{ruas: []models.Rua{{
		ID: ruaID, NomeRua: "RUA TESTE",
		Geometria: `{"type":"MultiLineString","coordinates":[[[-41.4,-21.8],[-41.2,-21.6]]]}`,
	}}}
	cache := &geocodificacaoCacheFake{}
	geocodificador := &geocodificadorFake{coordenada: &CoordenadaReferencia{Latitude: -20, Longitude: -40}}
	resolver := NewCoordenadaResolver(ruAs, cache, geocodificador)

	resultado, err := resolver.Resolver(context.Background(), SolicitacaoCoordenada{
		ChaveAgrupamento: "rua:10", NomeRua: "RUA TESTE", RuaID: &ruaID, PermitirExterno: true,
	})
	if err != nil {
		t.Fatalf("Resolver() erro inesperado: %v", err)
	}
	if resultado == nil || math.Abs(resultado.Latitude-(-21.7)) > 1e-9 || math.Abs(resultado.Longitude-(-41.3)) > 1e-9 {
		t.Fatalf("coordenada inesperada: %+v", resultado)
	}
	if len(resultado.Pontos) != 2 || resultado.Pontos[0].Latitude != -21.8 || resultado.Pontos[1].Longitude != -41.2 {
		t.Fatalf("pontos do traçado não foram preservados: %+v", resultado.Pontos)
	}
	if resultado.Fonte != models.FonteCoordenadaGeometria || geocodificador.chamadas != 0 {
		t.Fatalf("fonte=%q chamadas externas=%d", resultado.Fonte, geocodificador.chamadas)
	}
}

func TestCoordenadaResolverAgrupaGeometriasDosTrechos(t *testing.T) {
	ruAs := coordenadaRuaRepoFake{ruas: []models.Rua{
		{ID: 1, NomeRua: "AVENIDA TESTE - ATÉ 100 - LADO PAR", Geometria: `{"type":"Point","coordinates":[-41.4,-21.8]}`},
		{ID: 2, NomeRua: "AVENIDA TESTE - DE 101 AO FIM - LADO ÍMPAR", Geometria: `{"type":"Point","coordinates":[-41.2,-21.6]}`},
	}}
	resolver := NewCoordenadaResolver(ruAs, &geocodificacaoCacheFake{}, nil)

	resultado, err := resolver.Resolver(context.Background(), SolicitacaoCoordenada{
		ChaveAgrupamento: "nome:TESTE", NomeRua: "AVENIDA TESTE",
	})
	if err != nil || resultado == nil {
		t.Fatalf("Resolver() = %+v, erro=%v", resultado, err)
	}
	if math.Abs(resultado.Latitude-(-21.7)) > 1e-9 || math.Abs(resultado.Longitude-(-41.3)) > 1e-9 {
		t.Fatalf("coordenada inesperada: %+v", resultado)
	}
	if len(resultado.Pontos) != 2 {
		t.Fatalf("pontos dos trechos não foram preservados: %+v", resultado.Pontos)
	}
}

func TestCoordenadaResolverCacheiaFallbackExterno(t *testing.T) {
	cache := &geocodificacaoCacheFake{}
	geocodificador := &geocodificadorFake{coordenada: &CoordenadaReferencia{
		Latitude: -21.75, Longitude: -41.32, Fonte: models.FonteCoordenadaNominatim,
	}}
	resolver := NewCoordenadaResolver(coordenadaRuaRepoFake{}, cache, geocodificador)
	solicitacao := SolicitacaoCoordenada{ChaveAgrupamento: "rua:20", NomeRua: "RUA SEM MAPA", PermitirExterno: true}

	for i := 0; i < 2; i++ {
		resultado, err := resolver.Resolver(context.Background(), solicitacao)
		if err != nil || resultado == nil {
			t.Fatalf("Resolver() = %+v, erro=%v", resultado, err)
		}
	}
	if geocodificador.chamadas != 1 || cache.quantidade != 1 {
		t.Fatalf("chamadas externas=%d gravações cache=%d", geocodificador.chamadas, cache.quantidade)
	}
}

func TestCoordenadaResolverCacheiaResultadoNaoEncontrado(t *testing.T) {
	cache := &geocodificacaoCacheFake{}
	geocodificador := &geocodificadorFake{}
	resolver := NewCoordenadaResolver(coordenadaRuaRepoFake{}, cache, geocodificador)
	solicitacao := SolicitacaoCoordenada{ChaveAgrupamento: "rua:30", NomeRua: "RUA SEM RESULTADO", PermitirExterno: true}

	for i := 0; i < 2; i++ {
		resultado, err := resolver.Resolver(context.Background(), solicitacao)
		if err != nil || resultado != nil {
			t.Fatalf("Resolver() = %+v, erro=%v", resultado, err)
		}
	}
	if geocodificador.chamadas != 1 || cache.quantidade != 1 || cache.item.Encontrada {
		t.Fatalf("cache negativo inesperado: chamadas=%d gravações=%d item=%+v", geocodificador.chamadas, cache.quantidade, cache.item)
	}
}
