package models

import "time"

// CategoriaObservacao são as categorias fixas combinadas — texto livre
// dentro de uma categoria fechada, pra ficar fácil de filtrar/ler rápido
// no meio de uma entrega, sem virar bagunça de texto solto.
type CategoriaObservacao string

const (
	ObsAcesso          CategoriaObservacao = "acesso"
	ObsSeguranca       CategoriaObservacao = "seguranca"
	ObsNumeroIrregular CategoriaObservacao = "numero_irregular"
	ObsVariosNomes     CategoriaObservacao = "varios_nomes"
	ObsOutros          CategoriaObservacao = "outros"
)

// RuaObservacao é conhecimento de campo registrado por um admin sobre uma
// rua específica — o tipo de coisa que só quem já entregou ali sabe (rua
// sem saída, cachorro solto, numeração fora de ordem). Alimenta o Zé Rota
// além do que já está estruturado no cadastro de ruas.
type RuaObservacao struct {
	ID        uint                `gorm:"primaryKey;autoIncrement" json:"id"`
	RuaID     uint                `gorm:"not null;index" json:"rua_id"`
	Categoria CategoriaObservacao `gorm:"type:varchar(30);not null" json:"categoria"`
	Texto     string              `gorm:"type:text;not null" json:"texto"`
	CriadoPor string              `gorm:"type:varchar(20)" json:"criado_por"`
	CreatedAt time.Time           `json:"created_at"`
}

func (RuaObservacao) TableName() string {
	return "rua_observacoes"
}
