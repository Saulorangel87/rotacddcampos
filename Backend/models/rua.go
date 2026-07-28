package models

import (
	"time"
)

// Rua representa uma rua cadastrada no sistema de rotas de entrega.
// Futuramente pode ser estendida com campos geométricos (PostGIS).
type Rua struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	NomeRua     string    `gorm:"type:varchar(255);not null;index" json:"nome_rua"`
	Bairro      string    `gorm:"type:varchar(100)" json:"bairro"`
	CEP         string    `gorm:"type:varchar(20);index" json:"cep"`
	Distrito    string    `gorm:"type:varchar(100);not null;index" json:"distrito"`
	Rota        string    `gorm:"type:varchar(50);index" json:"rota"`
	Observacao  string    `gorm:"type:text" json:"observacao"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
