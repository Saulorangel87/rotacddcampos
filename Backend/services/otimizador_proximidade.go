package services

import "math"

// ParadaParaOtimizar é a entrada mínima para o motor local de proximidade.
// Ele não conhece trânsito ou malha viária: usa apenas distância geográfica.
type ParadaParaOtimizar struct {
	Chave           string
	NomeRua         string
	Quantidade      int
	Latitude        float64
	Longitude       float64
	FonteCoordenada string
	// Pontos contém vértices do traçado real da rua. Quando disponível, a
	// distância usada pelo motor é a menor distância até esses pontos, em vez
	// da distância até um centro médio que pode cair fora do traçado.
	Pontos []PontoGeografico
}

type OtimizadorRota interface {
	Otimizar(origem PontoGeografico, paradas []ParadaParaOtimizar) []ParadaParaOtimizar
}

type otimizadorProximidade struct{}

func NewOtimizadorProximidade() OtimizadorRota {
	return otimizadorProximidade{}
}

func (otimizadorProximidade) Otimizar(origem PontoGeografico, paradas []ParadaParaOtimizar) []ParadaParaOtimizar {
	if len(paradas) == 0 {
		return append([]ParadaParaOtimizar(nil), paradas...)
	}

	return rotaVizinhoMaisProximo(origem, paradas)
}

func rotaVizinhoMaisProximo(origem PontoGeografico, paradas []ParadaParaOtimizar) []ParadaParaOtimizar {
	pendentes := append([]ParadaParaOtimizar(nil), paradas...)
	resultado := make([]ParadaParaOtimizar, 0, len(paradas))
	atual := origem
	for len(pendentes) > 0 {
		indiceProxima := 0
		_, menorDistancia := pontoMaisProximo(atual, pendentes[0])
		for indice := 1; indice < len(pendentes); indice++ {
			_, distancia := pontoMaisProximo(atual, pendentes[indice])
			if distancia < menorDistancia || (distancia == menorDistancia && paradaVemAntes(pendentes[indice], pendentes[indiceProxima])) {
				indiceProxima = indice
				menorDistancia = distancia
			}
		}
		proxima := pendentes[indiceProxima]
		pontoEscolhido, _ := pontoMaisProximo(atual, proxima)
		// A parada persistida continua tendo uma coordenada simples para a API
		// e para a interface; ela representa o ponto do traçado usado nesta
		// transição. Os Pontos permanecem disponíveis apenas durante o cálculo.
		proxima.Latitude = pontoEscolhido.Latitude
		proxima.Longitude = pontoEscolhido.Longitude
		resultado = append(resultado, proxima)
		atual = pontoEscolhido
		pendentes = append(pendentes[:indiceProxima], pendentes[indiceProxima+1:]...)
	}
	return resultado
}

func paradaVemAntes(a, b ParadaParaOtimizar) bool {
	if a.NomeRua != b.NomeRua {
		return a.NomeRua < b.NomeRua
	}
	return a.Chave < b.Chave
}

func pontoDaParada(parada ParadaParaOtimizar) PontoGeografico {
	return PontoGeografico{Latitude: parada.Latitude, Longitude: parada.Longitude}
}

func pontoMaisProximo(origem PontoGeografico, parada ParadaParaOtimizar) (PontoGeografico, float64) {
	pontos := parada.Pontos
	if len(pontos) == 0 {
		ponto := pontoDaParada(parada)
		return ponto, distanciaHaversine(origem, ponto)
	}

	melhor := pontos[0]
	menor := distanciaHaversine(origem, melhor)
	for _, ponto := range pontos[1:] {
		distancia := distanciaHaversine(origem, ponto)
		if distancia < menor || (distancia == menor && pontoVemAntes(ponto, melhor)) {
			melhor = ponto
			menor = distancia
		}
	}
	return melhor, menor
}

func pontoVemAntes(a, b PontoGeografico) bool {
	if a.Latitude != b.Latitude {
		return a.Latitude < b.Latitude
	}
	return a.Longitude < b.Longitude
}

// distanciaHaversine retorna a distância em quilômetros entre dois pontos.
func distanciaHaversine(a, b PontoGeografico) float64 {
	const raioTerraKM = 6371.0088
	latitudeA := a.Latitude * math.Pi / 180
	latitudeB := b.Latitude * math.Pi / 180
	deltaLatitude := (b.Latitude - a.Latitude) * math.Pi / 180
	deltaLongitude := (b.Longitude - a.Longitude) * math.Pi / 180

	h := math.Sin(deltaLatitude/2)*math.Sin(deltaLatitude/2) +
		math.Cos(latitudeA)*math.Cos(latitudeB)*math.Sin(deltaLongitude/2)*math.Sin(deltaLongitude/2)
	return 2 * raioTerraKM * math.Atan2(math.Sqrt(h), math.Sqrt(1-h))
}
