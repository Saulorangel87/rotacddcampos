package repositories

import (
	"context"

	"github.com/empresa/rotas-entrega/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GeocodificacaoRepository interface {
	FindByChave(ctx context.Context, chave string) (*models.GeocodificacaoRua, error)
	Save(ctx context.Context, geocodificacao *models.GeocodificacaoRua) error
}

type geocodificacaoRepository struct {
	db *gorm.DB
}

func NewGeocodificacaoRepository(db *gorm.DB) GeocodificacaoRepository {
	return &geocodificacaoRepository{db: db}
}

func (r *geocodificacaoRepository) FindByChave(ctx context.Context, chave string) (*models.GeocodificacaoRua, error) {
	var geocodificacao models.GeocodificacaoRua
	err := r.db.WithContext(ctx).Where("chave_agrupamento = ?", chave).First(&geocodificacao).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &geocodificacao, err
}

func (r *geocodificacaoRepository) Save(ctx context.Context, geocodificacao *models.GeocodificacaoRua) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "chave_agrupamento"}},
		DoUpdates: clause.AssignmentColumns([]string{"nome_rua", "latitude", "longitude", "fonte", "encontrada", "updated_at"}),
	}).Create(geocodificacao).Error
}
