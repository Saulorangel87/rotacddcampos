package services

// Coordenadas fixas do CDD Campos dos Goytacazes, na Av. Sete de Setembro,
// 342. Este é o ponto inicial comum para clima e ordenamento.
const (
	LatitudeCDD  = -21.7545
	LongitudeCDD = -41.3244
)

type PontoGeografico struct {
	Latitude  float64
	Longitude float64
}

func PontoPartidaCDD() PontoGeografico {
	return PontoGeografico{Latitude: LatitudeCDD, Longitude: LongitudeCDD}
}
