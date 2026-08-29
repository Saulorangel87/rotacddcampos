package models

import "time"

// PlanoRedistritamentoRua é uma rua "órfã" dentro de um plano — ou seja, uma
// rua que hoje está num distrito marcado pra extinção e precisa ser
// realocada manualmente pra um distrito sobrevivente antes do plano poder
// ser aplicado. DistritoDestino fica vazio até o admin escolher.
type PlanoRedistritamentoRua struct {
	ID              uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	PlanoID         uint   `gorm:"not null;index" json:"plano_id"`
	RuaID           uint   `gorm:"not null;index" json:"rua_id"`
	NomeRua         string `gorm:"type:varchar(255)" json:"nome_rua"`
	Bairro          string `gorm:"type:varchar(100)" json:"bairro"`
	DistritoOrigem  string `gorm:"type:varchar(10)" json:"distrito_origem"`
	DistritoDestino string `gorm:"type:varchar(10)" json:"distrito_destino"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (PlanoRedistritamentoRua) TableName() string {
	return "plano_redistritamento_ruas"
}
