package repositories

import (
	"context"

	"github.com/empresa/rotas-entrega/models"
	"gorm.io/gorm"
)

type ParadaOrdenamentoRepository interface {
	ListByOrdenamento(ctx context.Context, ordenamentoID uint) ([]models.ParadaOrdenamento, error)
	ReplaceByOrdenamento(ctx context.Context, ordenamentoID uint, paradas []models.ParadaOrdenamento) error
	DeleteByOrdenamento(ctx context.Context, ordenamentoID uint) error
}

type paradaOrdenamentoRepository struct {
	db *gorm.DB
}

func NewParadaOrdenamentoRepository(db *gorm.DB) ParadaOrdenamentoRepository {
	return &paradaOrdenamentoRepository{db: db}
}

func (r *paradaOrdenamentoRepository) ListByOrdenamento(ctx context.Context, ordenamentoID uint) ([]models.ParadaOrdenamento, error) {
	var paradas []models.ParadaOrdenamento
	err := r.db.WithContext(ctx).Where("ordenamento_id = ?", ordenamentoID).Order("ordem_sugerida ASC").Find(&paradas).Error
	return paradas, err
}

func (r *paradaOrdenamentoRepository) ReplaceByOrdenamento(ctx context.Context, ordenamentoID uint, paradas []models.ParadaOrdenamento) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("ordenamento_id = ?", ordenamentoID).Delete(&models.ParadaOrdenamento{}).Error; err != nil {
			return err
		}
		if len(paradas) == 0 {
			return nil
		}
		return tx.Create(&paradas).Error
	})
}

func (r *paradaOrdenamentoRepository) DeleteByOrdenamento(ctx context.Context, ordenamentoID uint) error {
	return r.db.WithContext(ctx).Where("ordenamento_id = ?", ordenamentoID).Delete(&models.ParadaOrdenamento{}).Error
}
