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

func TestOtimizadorProximidadeEscolheMaisProximaAposCadaParada(t *testing.T) {
	origem := PontoGeografico{Latitude: -21.7545, Longitude: -41.3244}
	paradas := []ParadaParaOtimizar{
		{Chave: "he", NomeRua: "RUA HEMETERIO MARTINS", Latitude: -21.74741348, Longitude: -41.31991077},
		{Chave: "vi", NomeRua: "RUA VICTOR SENCE", Latitude: -21.74585041, Longitude: -41.31913047},
		{Chave: "ar", NomeRua: "RUA ARAÚJO SILVA", Latitude: -21.74483195, Longitude: -41.32411008},
		{Chave: "ad", NomeRua: "RUA ADVALDO MACIEL", Latitude: -21.74268008, Longitude: -41.32365472},
		{Chave: "hu", NomeRua: "RUA HUMBERTO DE CAMPOS", Latitude: -21.73931632, Longitude: -41.32429478},
		{Chave: "co", NomeRua: "RUA CORONEL WALTER KRAMER", Latitude: -21.73942655, Longitude: -41.32516038},
		{Chave: "li", NomeRua: "RUA LINDOLFO FRAGA", Latitude: -21.73648282, Longitude: -41.33158173},
		{Chave: "na", NomeRua: "RUA NAZÁRIO PEREIRA GOMES", Latitude: -21.73136348, Longitude: -41.32864929},
		{Chave: "al", NomeRua: "RUA ALCIDES VIEIRA MACIEL", Latitude: -21.72217914, Longitude: -41.30780713},
		{Chave: "sa", NomeRua: "RUA SANTO ANTÔNIO", Latitude: -21.76680614, Longitude: -41.30097881},
	}

	resultado := NewOtimizadorProximidade().Otimizar(origem, paradas)
	esperado := []string{"he", "vi", "ar", "ad", "hu", "co", "li", "na", "al", "sa"}
	for indice, chave := range esperado {
		if resultado[indice].Chave != chave {
			t.Fatalf("parada %d = %q, esperado %q (vizinho mais próximo)", indice+1, resultado[indice].Chave, chave)
		}
	}
}

func TestOtimizadorProximidadeUsaPontoMaisProximoDoTracado(t *testing.T) {
	origem := PontoGeografico{Latitude: 0, Longitude: 0}
	paradas := []ParadaParaOtimizar{
		{
			Chave: "rua:longe", NomeRua: "RUA LONGE", Latitude: 0.5, Longitude: 0.5,
			Pontos: []PontoGeografico{{Latitude: 0.5, Longitude: 0.5}, {Latitude: 0.01, Longitude: 0}},
		},
		{Chave: "rua:perto", NomeRua: "RUA PERTO", Latitude: 0.02, Longitude: 0},
	}

	resultado := NewOtimizadorProximidade().Otimizar(origem, paradas)
	if resultado[0].Chave != "rua:longe" {
		t.Fatalf("a rua deve ser comparada pelo ponto mais próximo do traçado: %q", resultado[0].Chave)
	}
	if math.Abs(resultado[0].Latitude-0.01) > 1e-9 || math.Abs(resultado[0].Longitude) > 1e-9 {
		t.Fatalf("o ponto escolhido da transição não foi persistido: %+v", resultado[0])
	}
	if resultado[1].Chave != "rua:perto" {
		t.Fatalf("segunda parada inesperada: %q", resultado[1].Chave)
	}
}

func TestOtimizadorProximidadeMantemVizinhoMaisProximoEmCadaTransicao(t *testing.T) {
	origem := PontoGeografico{Latitude: -21.7545, Longitude: -41.3244}
	paradas := []ParadaParaOtimizar{
		{Chave: "a", NomeRua: "RUA A", Latitude: -21.75, Longitude: -41.32, Pontos: []PontoGeografico{{Latitude: -21.75, Longitude: -41.32}, {Latitude: -21.751, Longitude: -41.321}}},
		{Chave: "b", NomeRua: "RUA B", Latitude: -21.76, Longitude: -41.31, Pontos: []PontoGeografico{{Latitude: -21.76, Longitude: -41.31}, {Latitude: -21.755, Longitude: -41.31}}},
		{Chave: "c", NomeRua: "RUA C", Latitude: -21.77, Longitude: -41.33, Pontos: []PontoGeografico{{Latitude: -21.77, Longitude: -41.33}, {Latitude: -21.765, Longitude: -41.329}}},
	}

	resultado := NewOtimizadorProximidade().Otimizar(origem, paradas)
	pendentes := append([]ParadaParaOtimizar(nil), paradas...)
	atual := origem
	for indice, escolhida := range resultado {
		_, distanciaEscolhida := pontoMaisProximo(atual, escolhida)
		for _, candidata := range pendentes {
			_, distanciaCandidata := pontoMaisProximo(atual, candidata)
			if distanciaEscolhida > distanciaCandidata+1e-9 {
				t.Fatalf("transição %d escolheu %.9f km, mas havia %.9f km disponíveis", indice+1, distanciaEscolhida, distanciaCandidata)
			}
		}
		ponto, _ := pontoMaisProximo(atual, escolhida)
		atual = ponto
		for i, candidata := range pendentes {
			if candidata.Chave == escolhida.Chave {
				pendentes = append(pendentes[:i], pendentes[i+1:]...)
				break
			}
		}
	}
}
