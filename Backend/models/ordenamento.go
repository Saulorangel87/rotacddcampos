package models

import "time"

const StatusOrdenamentoEmAndamento = "em_andamento"

// Ordenamento representa uma sessão de organização de encomendas de um usuário.
// Objetos e paradas serão relacionados a esta entidade nas próximas fases.
type Ordenamento struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UsuarioID uint      `gorm:"not null;uniqueIndex:idx_ordenamentos_usuario_ativo,where:status = 'em_andamento'" json:"usuario_id"`
	Status    string    `gorm:"type:varchar(20);not null;default:'em_andamento';index" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Ordenamento) TableName() string {
	return "ordenamentos"
}
