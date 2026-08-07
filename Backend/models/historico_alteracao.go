package models

import "time"

// HistoricoAlteracao registra toda movimentação sensível do sistema — hoje
// dois tipos: "rua" (rua mudou de distrito) e "folga" (crédito/débito de
// folga lançado ou excluído). Criado automaticamente pelos services
// correspondentes, nunca chamado direto pelo front.
type HistoricoAlteracao struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Tipo            string    `gorm:"type:varchar(20);not null;default:'rua'" json:"tipo"`
	RuaID           uint      `gorm:"index" json:"rua_id,omitempty"`
	NomeRua         string    `gorm:"type:varchar(255)" json:"nome_rua,omitempty"`
	DistritoOrigem  string    `gorm:"type:varchar(10)" json:"distrito_origem,omitempty"`
	DistritoDestino string    `gorm:"type:varchar(10)" json:"distrito_destino,omitempty"`
	Descricao       string    `gorm:"type:varchar(255)" json:"descricao,omitempty"`
	Usuario         string    `gorm:"type:varchar(150)" json:"usuario"`
	CreatedAt       time.Time `json:"data"`
}

func (HistoricoAlteracao) TableName() string {
	return "historico_alteracoes"
}
