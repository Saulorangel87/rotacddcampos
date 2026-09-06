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

func TestOtimizadorProximidadeDesempataParadasIguaisPelaChave(t *testing.T) {
	origem := PontoGeografico{}
	paradas := []ParadaParaOtimizar{
		{Chave: "rua:20", NomeRua: "AVENIDA SETE DE SETEMBRO", Latitude: 0.01, Longitude: 0},
		{Chave: "rua:10", NomeRua: "AVENIDA SETE DE SETEMBRO", Latitude: 0.01, Longitude: 0},
	}

	resultado := NewOtimizadorProximidade().Otimizar(origem, paradas)
	if resultado[0].Chave != "rua:10" {
		t.Fatalf("desempate = %q, esperado rua:10", resultado[0].Chave)
	}
}

func TestOtimizadorProximidadeMantemPrimeiraParadaMaisProximaAposDoisOpt(t *testing.T) {
	origem := PontoGeografico{}
	paradas := []ParadaParaOtimizar{
		{Chave: "a", NomeRua: "RUA A", Latitude: 2.28265, Longitude: -1.53681},
		{Chave: "b", NomeRua: "RUA B", Latitude: 2.85567, Longitude: -4.72631},
		{Chave: "c", NomeRua: "RUA C", Latitude: 2.55204, Longitude: 6.09974},
		{Chave: "d", NomeRua: "RUA D", Latitude: 1.25306, Longitude: 8.64923},
		{Chave: "e", NomeRua: "RUA E", Latitude: -5.24234, Longitude: 9.96405},
		{Chave: "f", NomeRua: "RUA F", Latitude: -7.98961, Longitude: 8.91075},
	}

	resultado := NewOtimizadorProximidade().Otimizar(origem, paradas)
	if resultado[0].Chave != "a" {
		t.Fatalf("primeira parada = %q, esperado a (mais próxima do CDD)", resultado[0].Chave)
	}
}

func TestOtimizadorProximidadeNaoAumentaRotaInicial(t *testing.T) {
	origem := PontoGeografico{}
	paradas := []ParadaParaOtimizar{
		{Chave: "a", Latitude: 0.02, Longitude: 0.01},
		{Chave: "b", Latitude: 0.04, Longitude: 0.04},
		{Chave: "c", Latitude: 0.01, Longitude: 0.04},
		{Chave: "d", Latitude: 0.05, Longitude: 0.01},
		{Chave: "e", Latitude: 0.03, Longitude: 0.06},
	}

	inicial := rotaVizinhoMaisProximo(origem, paradas)
	final := NewOtimizadorProximidade().Otimizar(origem, paradas)
	if depois, antes := distanciaDaSequencia(origem, final), distanciaDaSequencia(origem, inicial); depois > antes+0.000001 {
		t.Fatalf("2-opt restrito aumentou a distância: antes=%f depois=%f", antes, depois)
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
