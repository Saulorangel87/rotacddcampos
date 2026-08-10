package repositories

import (
	"context"

	"github.com/empresa/rotas-entrega/models"
	"gorm.io/gorm"
)

type RuaObservacaoRepository interface {
	FindByRuaID(ctx context.Context, ruaID uint) ([]models.RuaObservacao, error)
	Create(ctx context.Context, obs *models.RuaObservacao) error
	FindByID(ctx context.Context, id uint) (*models.RuaObservacao, error)
	Delete(ctx context.Context, id uint) error
}

type ruaObservacaoRepository struct {
	db *gorm.DB
}

func NewRuaObservacaoRepository(db *gorm.DB) RuaObservacaoRepository {
	return &ruaObservacaoRepository{db: db}
}

func (r *ruaObservacaoRepository) FindByRuaID(ctx context.Context, ruaID uint) ([]models.RuaObservacao, error) {
	var obs []models.RuaObservacao
	err := r.db.WithContext(ctx).
		Where("rua_id = ?", ruaID).
		Order("created_at desc").
		Find(&obs).Error
	return obs, err
}

func (r *ruaObservacaoRepository) Create(ctx context.Context, obs *models.RuaObservacao) error {
	return r.db.WithContext(ctx).Create(obs).Error
}

func (r *ruaObservacaoRepository) FindByID(ctx context.Context, id uint) (*models.RuaObservacao, error) {
	var obs models.RuaObservacao
	if err := r.db.WithContext(ctx).First(&obs, id).Error; err != nil {
		return nil, err
	}
	return &obs, nil
}

func (r *ruaObservacaoRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.RuaObservacao{}, id).Error
}
