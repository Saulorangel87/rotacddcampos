package database

import (
	"github.com/empresa/rotas-entrega/models"
	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) error {
	// unaccent deixa a busca de rua tolerante a acento (ex: digitar "araujo"
	// acha "Araújo") — sem isso, ILIKE compara caractere a caractere e falha
	// em qualquer diferença de acentuação, mesmo com o nome certo.
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS unaccent").Error; err != nil {
		return err
	}

	return db.AutoMigrate(
		&models.Rua{},
		&models.Colaborador{},
		&models.Distrito{},
		&models.HistoricoAlteracao{},
		&models.Usuario{},
		&models.AcessoLogin{},
		&models.FolgaLancamento{},
		&models.RuaObservacao{},
	)
}
