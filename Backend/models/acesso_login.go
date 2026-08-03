package models

import "time"

// AcessoLogin registra toda tentativa de login (sucesso ou falha), pra auditoria
// e pra detectar tentativa de uso indevido de matrícula de terceiros.
type AcessoLogin struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Matricula string    `gorm:"type:varchar(20);not null;index" json:"matricula"`
	Sucesso   bool      `gorm:"not null" json:"sucesso"`
	IP        string    `gorm:"type:varchar(45)" json:"ip"`
	CreatedAt time.Time `json:"data"`
}

func (AcessoLogin) TableName() string {
	return "acessos_login"
}
