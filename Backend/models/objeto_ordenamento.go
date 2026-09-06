package models

import "time"

const (
	OrigemEntradaManual  = "manual"
	OrigemEntradaVoz     = "voz"
	OrigemEntradaScanner = "scanner"

	StatusResolucaoIdentificado = "identificado"
	StatusResolucaoPendente     = "pendente"
)

// ObjetoOrdenamento guarda somente os dados operacionais necessários para
// identificar e agrupar uma encomenda. Não armazena dados do destinatário.
type ObjetoOrdenamento struct {
	ID               uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrdenamentoID    uint      `gorm:"not null;index" json:"ordenamento_id"`
	RuaID            *uint     `gorm:"index" json:"rua_id,omitempty"`
	TextoEntrada     string    `gorm:"type:varchar(255);not null" json:"texto_entrada"`
	NomeRua          string    `gorm:"type:varchar(255)" json:"nome_rua,omitempty"`
	ChaveAgrupamento string    `gorm:"type:varchar(255);index" json:"chave_agrupamento,omitempty"`
	Numero           string    `gorm:"type:varchar(30)" json:"numero,omitempty"`
	CEP              string    `gorm:"type:varchar(20)" json:"cep,omitempty"`
	Latitude         *float64  `json:"latitude,omitempty"`
	Longitude        *float64  `json:"longitude,omitempty"`
	FonteCoordenada  string    `gorm:"type:varchar(30)" json:"fonte_coordenada,omitempty"`
	OrigemEntrada    string    `gorm:"type:varchar(20);not null" json:"origem_entrada"`
	StatusResolucao  string    `gorm:"type:varchar(20);not null;index" json:"status_resolucao"`
	MotivoPendencia  string    `gorm:"type:varchar(50)" json:"motivo_pendencia,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (ObjetoOrdenamento) TableName() string {
	return "objetos_ordenamento"
}
