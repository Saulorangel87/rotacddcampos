package models

import "time"

// Distrito guarda os metadados de cada distrito (601-624...), incluindo o
// GeoJSON do polígono quando disponível. Usa o próprio código do distrito
// como chave primária (o mesmo texto já usado em ruas.distrito), evitando
// criar um id novo e ter que migrar as 2264 ruas existentes.
type Distrito struct {
	Codigo    string    `gorm:"primaryKey;type:varchar(10)" json:"codigo"`
	Nome      string    `gorm:"type:varchar(100)" json:"nome"`
	Cor       string    `gorm:"type:varchar(20)" json:"cor"`
	GeoJSON   string    `gorm:"type:text" json:"geojson"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Distrito) TableName() string {
	return "distritos"
}
