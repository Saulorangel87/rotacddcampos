package repositories

import (
	"context"

	"github.com/empresa/rotas-entrega/models"
	"gorm.io/gorm"
)

type HistoricoRepository interface {
	Create(ctx context.Context, registro *models.HistoricoAlteracao) error
	FindAll(ctx context.Context, pagina, limite int) ([]models.HistoricoAlteracao, int64, error)
}

type historicoRepository struct {
	db *gorm.DB
}

func NewHistoricoRepository(db *gorm.DB) HistoricoRepository {
	return &historicoRepository{db: db}
}

func (r *historicoRepository) Create(ctx context.Context, registro *models.HistoricoAlteracao) error {
	return r.db.WithContext(ctx).Create(registro).Error
}

func (r *historicoRepository) FindAll(ctx context.Context, pagina, limite int) ([]models.HistoricoAlteracao, int64, error) {
	if pagina < 1 {
		pagina = 1
	}
	if limite < 1 || limite > 100 {
		limite = 10
	}

	var total int64
	if err := r.db.WithContext(ctx).Model(&models.HistoricoAlteracao{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var registros []models.HistoricoAlteracao
	offset := (pagina - 1) * limite
	err := r.db.WithContext(ctx).
		Order("created_at desc").
		Limit(limite).
		Offset(offset).
		Find(&registros).Error

	return registros, total, err
}
