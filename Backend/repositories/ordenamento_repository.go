package repositories

import (
	"context"

	"github.com/empresa/rotas-entrega/models"
	"gorm.io/gorm"
)

type OrdenamentoRepository interface {
	FindAtivoByUsuario(ctx context.Context, usuarioID uint) (*models.Ordenamento, error)
	Create(ctx context.Context, ordenamento *models.Ordenamento) error
}

type ordenamentoRepository struct {
	db *gorm.DB
}

func NewOrdenamentoRepository(db *gorm.DB) OrdenamentoRepository {
	return &ordenamentoRepository{db: db}
}

func (r *ordenamentoRepository) FindAtivoByUsuario(ctx context.Context, usuarioID uint) (*models.Ordenamento, error) {
	var ordenamento models.Ordenamento
	err := r.db.WithContext(ctx).
		Where("usuario_id = ? AND status = ?", usuarioID, models.StatusOrdenamentoEmAndamento).
		Order("created_at DESC").
		First(&ordenamento).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &ordenamento, nil
}

func (r *ordenamentoRepository) Create(ctx context.Context, ordenamento *models.Ordenamento) error {
	return r.db.WithContext(ctx).Create(ordenamento).Error
}
