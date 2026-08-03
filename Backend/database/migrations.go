package database

import (
	"github.com/empresa/rotas-entrega/models"
	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Rua{},
		&models.Colaborador{},
		&models.Distrito{},
		&models.HistoricoAlteracao{},
		&models.Usuario{},
		&models.AcessoLogin{},
	)
}
