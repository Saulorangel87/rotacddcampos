package models

import (
	"time"
)

// Colaborador representa um funcionário cadastrado na tabela `colaboradores`,
// importada da planilha de efetivo do CDD Campos dos Goytacazes.
type Colaborador struct {
	ID             uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Matricula      string     `gorm:"type:varchar(20);not null" json:"matricula"`
	Nome           string     `gorm:"type:varchar(150);not null;index" json:"nome"`
	Funcao         string     `gorm:"type:varchar(50)" json:"funcao"`
	DataAdmissao   *time.Time `gorm:"type:date" json:"data_admissao"`
	DataNascimento *time.Time `gorm:"type:date" json:"data_nascimento"`
	Cargo          string     `gorm:"type:varchar(50)" json:"cargo"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// TableName força o nome certo da tabela (plural em português).
// Sem isso, o GORM usaria a regra de plural em inglês e criaria/buscaria
// em "colaboradors" em vez da tabela real "colaboradores".
func (Colaborador) TableName() string {
	return "colaboradores"
}
