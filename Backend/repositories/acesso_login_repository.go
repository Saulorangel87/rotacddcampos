package repositories

import (
	"context"

	"github.com/empresa/rotas-entrega/models"
	"gorm.io/gorm"
)

type AcessoLoginRepository interface {
	Create(ctx context.Context, acesso *models.AcessoLogin) error
}

type acessoLoginRepository struct {
	db *gorm.DB
}

func NewAcessoLoginRepository(db *gorm.DB) AcessoLoginRepository {
	return &acessoLoginRepository{db: db}
}

func (r *acessoLoginRepository) Create(ctx context.Context, acesso *models.AcessoLogin) error {
	return r.db.WithContext(ctx).Create(acesso).Error
}
