package services

import (
	"math"
	"testing"
)

func TestDistanciaHaversine(t *testing.T) {
	origem := PontoGeografico{Latitude: -21.7545, Longitude: -41.3244}
	if distancia := distanciaHaversine(origem, origem); distancia != 0 {
		t.Fatalf("distância do mesmo ponto = %f, esperado 0", distancia)
	}

	distancia := distanciaHaversine(PontoGeografico{}, PontoGeografico{Longitude: 1})
	if math.Abs(distancia-111.195) > 0.1 {
		t.Fatalf("distância de um grau no equador = %f km, esperado aproximadamente 111.195 km", distancia)
	}
}

func TestOtimizadorProximidadeComecaPelaRuaMaisProxima(t *testing.T) {
	origem := PontoGeografico{}
	paradas := []ParadaParaOtimizar{
		{Chave: "c", NomeRua: "RUA C", Latitude: 0.03, Longitude: 0},
		{Chave: "a", NomeRua: "RUA A", Latitude: 0.01, Longitude: 0},
		{Chave: "b", NomeRua: "RUA B", Latitude: 0.02, Longitude: 0},
	}

	resultado := NewOtimizadorProximidade().Otimizar(origem, paradas)
	if len(resultado) != len(paradas) {
		t.Fatalf("quantidade de paradas = %d, esperado %d", len(resultado), len(paradas))
	}
	for indice, chave := range []string{"a", "b", "c"} {
		if resultado[indice].Chave != chave {
			t.Fatalf("parada %d = %q, esperado %q", indice, resultado[indice].Chave, chave)
		}
	}
}

func TestMelhorarComDoisOptNaoAumentaDistancia(t *testing.T) {
	origem := PontoGeografico{}
	rota := []ParadaParaOtimizar{
		{NomeRua: "A", Latitude: 0.02, Longitude: 0.01},
		{NomeRua: "B", Latitude: 0.04, Longitude: 0.04},
		{NomeRua: "C", Latitude: 0.01, Longitude: 0.04},
		{NomeRua: "D", Latitude: 0.05, Longitude: 0.01},
	}
	antes := distanciaDaSequencia(origem, rota)
	depois := distanciaDaSequencia(origem, melhorarComDoisOpt(origem, rota))
	if depois > antes+0.000001 {
		t.Fatalf("2-opt aumentou distância: antes=%f depois=%f", antes, depois)
	}
}

func distanciaDaSequencia(origem PontoGeografico, rota []ParadaParaOtimizar) float64 {
	if len(rota) == 0 {
		return 0
	}
	total := distanciaHaversine(origem, pontoDaParada(rota[0]))
	for indice := 1; indice < len(rota); indice++ {
		total += distanciaHaversine(pontoDaParada(rota[indice-1]), pontoDaParada(rota[indice]))
	}
	return total
}
