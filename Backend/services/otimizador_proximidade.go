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

	pendentes := append([]ParadaParaOtimizar(nil), paradas...)
	resultado := make([]ParadaParaOtimizar, 0, len(paradas))
	atual := origem
	for len(pendentes) > 0 {
		indiceProxima := 0
		menorDistancia := distanciaHaversine(atual, pontoDaParada(pendentes[0]))
		for indice := 1; indice < len(pendentes); indice++ {
			distancia := distanciaHaversine(atual, pontoDaParada(pendentes[indice]))
			if distancia < menorDistancia || (distancia == menorDistancia && pendentes[indice].NomeRua < pendentes[indiceProxima].NomeRua) {
				indiceProxima = indice
				menorDistancia = distancia
			}
		}
		proxima := pendentes[indiceProxima]
		resultado = append(resultado, proxima)
		atual = pontoDaParada(proxima)
		pendentes = append(pendentes[:indiceProxima], pendentes[indiceProxima+1:]...)
	}
	return melhorarComDoisOpt(origem, resultado)
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

func melhorarComDoisOpt(origem PontoGeografico, rota []ParadaParaOtimizar) []ParadaParaOtimizar {
	if len(rota) < 4 {
		return rota
	}
	melhorou := true
	for melhorou {
		melhorou = false
		for inicio := 0; inicio < len(rota)-2; inicio++ {
			antes := origem
			if inicio > 0 {
				antes = pontoDaParada(rota[inicio-1])
			}
			for fim := inicio + 1; fim < len(rota)-1; fim++ {
				atual := pontoDaParada(rota[inicio])
				depois := pontoDaParada(rota[fim+1])
				custoAtual := distanciaHaversine(antes, atual) + distanciaHaversine(pontoDaParada(rota[fim]), depois)
				custoInvertido := distanciaHaversine(antes, pontoDaParada(rota[fim])) + distanciaHaversine(atual, depois)
				if custoInvertido+0.000001 < custoAtual {
					inverterTrecho(rota, inicio, fim)
					melhorou = true
				}
			}
		}
	}
	return rota
}

func inverterTrecho(valores []ParadaParaOtimizar, inicio, fim int) {
	for inicio < fim {
		valores[inicio], valores[fim] = valores[fim], valores[inicio]
		inicio++
		fim--
	}
}
