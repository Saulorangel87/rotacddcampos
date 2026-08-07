package repositories

import (
	"context"

	"github.com/empresa/rotas-entrega/models"
	"gorm.io/gorm"
)

type FolgaRepository interface {
	FindByMatricula(ctx context.Context, matricula string) ([]models.FolgaLancamento, error)
	Create(ctx context.Context, lancamento *models.FolgaLancamento) error
	FindByID(ctx context.Context, id uint) (*models.FolgaLancamento, error)
	Delete(ctx context.Context, id uint) error
}

type folgaRepository struct {
	db *gorm.DB
}

func NewFolgaRepository(db *gorm.DB) FolgaRepository {
	return &folgaRepository{db: db}
}

// FindByMatricula traz o extrato completo em ordem cronológica (mais antigo
// primeiro), pra fazer sentido de ler como um extrato bancário.
func (r *folgaRepository) FindByMatricula(ctx context.Context, matricula string) ([]models.FolgaLancamento, error) {
	var lancamentos []models.FolgaLancamento
	err := r.db.WithContext(ctx).
		Where("matricula = ?", matricula).
		Order("data_referencia asc, created_at asc").
		Find(&lancamentos).Error
	return lancamentos, err
}

func (r *folgaRepository) Create(ctx context.Context, lancamento *models.FolgaLancamento) error {
	return r.db.WithContext(ctx).Create(lancamento).Error
}

func (r *folgaRepository) FindByID(ctx context.Context, id uint) (*models.FolgaLancamento, error) {
	var lancamento models.FolgaLancamento
	err := r.db.WithContext(ctx).First(&lancamento, id).Error
	if err != nil {
		return nil, err
	}
	return &lancamento, nil
}

func (r *folgaRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.FolgaLancamento{}, id).Error
}
