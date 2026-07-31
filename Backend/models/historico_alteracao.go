package models

import "time"

// HistoricoAlteracao registra toda vez que uma rua muda de distrito.
// Criado automaticamente pelo backend quando o distrito de uma rua muda
// (ver services/rua_service.go), não precisa ser chamado manualmente pelo front.
type HistoricoAlteracao struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RuaID           uint      `gorm:"not null;index" json:"rua_id"`
	NomeRua         string    `gorm:"type:varchar(255)" json:"nome_rua"`
	DistritoOrigem  string    `gorm:"type:varchar(10)" json:"distrito_origem"`
	DistritoDestino string    `gorm:"type:varchar(10)" json:"distrito_destino"`
	Usuario         string    `gorm:"type:varchar(150)" json:"usuario"`
	CreatedAt       time.Time `json:"data"`
}

func (HistoricoAlteracao) TableName() string {
	return "historico_alteracoes"
}
