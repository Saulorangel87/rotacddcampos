package models

import "time"

// Status possíveis de um PlanoRedistritamento.
const (
	StatusRedistritamentoRascunho  = "rascunho"
	StatusRedistritamentoConcluido = "concluido"
	StatusRedistritamentoAplicado  = "aplicado"
)

// PlanoRedistritamento representa uma "rodada" de redução (ou, no futuro,
// aumento) na quantidade de distritos. Enquanto o status não for "aplicado",
// nada muda de verdade no site — é só um rascunho de trabalho que o admin
// vai preenchendo aos poucos.
type PlanoRedistritamento struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`
	// Tipo hoje só suporta "reducao" — "aumento" fica reservado pro futuro.
	Tipo             string `gorm:"type:varchar(20);not null;default:'reducao'" json:"tipo"`
	Status           string `gorm:"type:varchar(20);not null;default:'rascunho'" json:"status"`
	QuantidadeAtual  int    `json:"quantidade_atual"`
	QuantidadeAlvo   int    `json:"quantidade_alvo"`
	// DistritosExtintos guarda os códigos que vão sumir, separados por vírgula
	// (ex: "618,619,620,621,622,623,624") — são os de maior código dentre os
	// ativos no momento em que o plano foi criado.
	DistritosExtintos string `gorm:"type:text" json:"distritos_extintos"`

	CriadoPor   string     `gorm:"type:varchar(150)" json:"criado_por"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ConcluidoEm *time.Time `json:"concluido_em,omitempty"`
	AplicadoEm  *time.Time `json:"aplicado_em,omitempty"`
	AplicadoPor string     `gorm:"type:varchar(150)" json:"aplicado_por,omitempty"`

	Ruas []PlanoRedistritamentoRua `gorm:"foreignKey:PlanoID" json:"ruas,omitempty"`
}

func (PlanoRedistritamento) TableName() string {
	return "planos_redistritamento"
}
