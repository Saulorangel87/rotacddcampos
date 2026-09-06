package models

import "time"

const (
	FonteCoordenadaGeometria = "geometria_interna"
	FonteCoordenadaNominatim = "nominatim"
)

// GeocodificacaoRua guarda somente o resultado externo de uma rua já
// identificada no cadastro interno. A chave evita repetir consultas ao provedor.
type GeocodificacaoRua struct {
	ID               uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ChaveAgrupamento string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"chave_agrupamento"`
	NomeRua          string    `gorm:"type:varchar(255);not null" json:"nome_rua"`
	Latitude         float64   `gorm:"not null" json:"latitude"`
	Longitude        float64   `gorm:"not null" json:"longitude"`
	Fonte            string    `gorm:"type:varchar(30);not null" json:"fonte"`
	Encontrada       bool      `gorm:"not null" json:"encontrada"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (GeocodificacaoRua) TableName() string {
	return "geocodificacoes_ruas"
}
