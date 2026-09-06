package models

import "time"

// ParadaOrdenamento representa uma rua agrupada e suas posições na sequência
// sugerida pelo algoritmo e, quando ajustada, na ordem final escolhida.
type ParadaOrdenamento struct {
	ID                uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrdenamentoID     uint      `gorm:"not null;uniqueIndex:idx_paradas_ordenamento_chave" json:"ordenamento_id"`
	ChaveAgrupamento  string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_paradas_ordenamento_chave" json:"chave_agrupamento"`
	NomeRua           string    `gorm:"type:varchar(255);not null" json:"nome_rua"`
	QuantidadeObjetos int       `gorm:"not null" json:"quantidade_objetos"`
	Latitude          float64   `gorm:"not null" json:"latitude"`
	Longitude         float64   `gorm:"not null" json:"longitude"`
	OrdemSugerida     int       `gorm:"not null" json:"ordem_sugerida"`
	OrdemFinal        *int      `json:"ordem_final,omitempty"`
	FonteOrdemFinal   string    `gorm:"type:varchar(20)" json:"fonte_ordem_final,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (ParadaOrdenamento) TableName() string {
	return "paradas_ordenamento"
}
