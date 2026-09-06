package models

import "time"

// CorrecaoOrdenamento preserva uma versão confirmada, independente das paradas
// temporárias da carga. Apenas Reutilizar/DesativadaEm controlam a referência
// pessoal; as sequências e a autoria não são alteradas após a criação.
type CorrecaoOrdenamento struct {
	ID             uint             `gorm:"primaryKey" json:"id"`
	UsuarioID      uint             `gorm:"not null;index:idx_correcoes_usuario_contexto,priority:1;uniqueIndex:idx_referencia_pessoal_ativa,where:reutilizar = true AND desativada_em IS NULL" json:"-"`
	OrdenamentoID  uint             `gorm:"not null" json:"ordenamento_id"`
	Contexto       string           `gorm:"type:varchar(150);not null;index:idx_correcoes_usuario_contexto,priority:2;uniqueIndex:idx_referencia_pessoal_ativa,where:reutilizar = true AND desativada_em IS NULL" json:"-"`
	AssinaturaRuas string           `gorm:"type:varchar(64);not null;index:idx_correcoes_usuario_contexto,priority:3;uniqueIndex:idx_referencia_pessoal_ativa,where:reutilizar = true AND desativada_em IS NULL" json:"-"`
	Reutilizar     bool             `gorm:"not null;default:false" json:"reutilizar"`
	Paradas        []ParadaCorrecao `gorm:"serializer:json;type:jsonb;not null" json:"paradas"`
	CreatedAt      time.Time        `json:"created_at"`
	DesativadaEm   *time.Time       `json:"desativada_em,omitempty"`
}

type ParadaCorrecao struct {
	ChaveAgrupamento string `json:"chave_agrupamento"`
	NomeRua          string `json:"nome_rua"`
	OrdemSugerida    int    `json:"ordem_sugerida"`
	OrdemFinal       int    `json:"ordem_final"`
}

func (CorrecaoOrdenamento) TableName() string { return "correcoes_ordenamento" }
