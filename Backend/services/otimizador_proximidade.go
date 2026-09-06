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
}

type OtimizadorRota interface {
	Otimizar(origem PontoGeografico, paradas []ParadaParaOtimizar) []ParadaParaOtimizar
}

type otimizadorProximidade struct{}

func NewOtimizadorProximidade() OtimizadorRota {
	return otimizadorProximidade{}
}

func (otimizadorProximidade) Otimizar(origem PontoGeografico, paradas []ParadaParaOtimizar) []ParadaParaOtimizar {
	if len(paradas) < 2 {
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
		menorDistancia := distanciaHaversine(atual, pontoDaParada(pendentes[0]))
		for indice := 1; indice < len(pendentes); indice++ {
			distancia := distanciaHaversine(atual, pontoDaParada(pendentes[indice]))
			if distancia < menorDistancia || (distancia == menorDistancia && paradaVemAntes(pendentes[indice], pendentes[indiceProxima])) {
				indiceProxima = indice
				menorDistancia = distancia
			}
		}
		proxima := pendentes[indiceProxima]
		resultado = append(resultado, proxima)
		atual = pontoDaParada(proxima)
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
