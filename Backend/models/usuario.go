package models

import "time"

// Usuario representa uma conta de acesso ao sistema, vinculada a um Colaborador.
// Papel controla o nível de acesso: "colaborador" (consulta autenticada) ou "admin" (edição).
type Usuario struct {
	ID               uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Matricula        string     `gorm:"type:varchar(20);not null;uniqueIndex" json:"matricula"`
	SenhaHash        string     `gorm:"type:varchar(100);not null" json:"-"`
	Papel            string     `gorm:"type:varchar(20);not null;default:colaborador" json:"papel"`
	ColaboradorID    *uint      `json:"colaborador_id"`
	SenhaProvisoria  bool       `gorm:"not null;default:true" json:"senha_provisoria"`
	TentativasFalhas int        `gorm:"not null;default:0" json:"-"`
	BloqueadoAte     *time.Time `json:"-"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func (Usuario) TableName() string {
	return "usuarios"
}

// PapelAdmin e PapelColaborador são os únicos papéis válidos hoje.
const (
	PapelAdmin       = "admin"
	PapelColaborador = "colaborador"
)
