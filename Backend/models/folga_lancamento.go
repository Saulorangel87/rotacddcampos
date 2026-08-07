package models

import "time"

// TipoFolga distingue um crédito (folga ganha, com justificativa) de um
// débito (folga efetivamente tirada).
type TipoFolga string

const (
	FolgaCredito TipoFolga = "credito"
	FolgaDebito  TipoFolga = "debito"
)

// FolgaLancamento é um lançamento no "livro-razão" de folgas de um
// colaborador. Nunca é sobrescrito: pra corrigir um erro, lança-se um novo
// registro ou exclui-se o errado (com auditoria de quem excluiu via
// historico_alteracoes, se algum dia precisar) — sempre com autor e data
// gravados, pra existir prova em caso de questionamento futuro.
//
// Saldo do colaborador = soma(quantidade onde tipo=credito) - soma(quantidade onde tipo=debito).
type FolgaLancamento struct {
	ID             uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Matricula      string     `gorm:"type:varchar(20);not null;index" json:"matricula"`
	Tipo           TipoFolga  `gorm:"type:varchar(10);not null" json:"tipo"`
	Quantidade     int        `gorm:"not null;default:1" json:"quantidade"`
	Motivo         string     `gorm:"type:varchar(255);not null" json:"motivo"`
	DataReferencia *time.Time `gorm:"type:date" json:"data_referencia"`
	CriadoPor      string     `gorm:"type:varchar(20)" json:"criado_por"`
	CreatedAt      time.Time  `json:"created_at"`
}

func (FolgaLancamento) TableName() string {
	return "folgas_lancamentos"
}
