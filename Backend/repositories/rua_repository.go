package repositories

import (
	"context"
	"github.com/empresa/rotas-entrega/models"
	"gorm.io/gorm"
)

type RuaRepository interface {
	FindAll(ctx context.Context, filters map[string]string) ([]models.Rua, error)
	FindByID(ctx context.Context, id uint) (*models.Rua, error)
	Create(ctx context.Context, rua *models.Rua) error
	Update(ctx context.Context, rua *models.Rua) error
	Delete(ctx context.Context, id uint) error
	ContarDistritosDistintos(ctx context.Context) (int64, error)
}

type ruaRepository struct {
	db *gorm.DB
}

func NewRuaRepository(db *gorm.DB) RuaRepository {
	return &ruaRepository{db: db}
}

func (r *ruaRepository) FindAll(ctx context.Context, filters map[string]string) ([]models.Rua, error) {
	var ruas []models.Rua
	query := r.db.WithContext(ctx).Model(&models.Rua{})

	if nome, ok := filters["nome"]; ok && nome != "" {
		// unaccent nos dois lados: sem acento no que a pessoa digitou
		// continua achando rua com acento no nome, e vice-versa.
		query = query.Where("unaccent(nome_rua) ILIKE unaccent(?)", "%"+nome+"%")
	}
	if cep, ok := filters["cep"]; ok && cep != "" {
		query = query.Where("cep LIKE ?", "%"+cep+"%")
	}
	if distrito, ok := filters["distrito"]; ok && distrito != "" {
		query = query.Where("distrito ILIKE ?", "%"+distrito+"%")
	}

	err := query.Order("nome_rua asc").Find(&ruas).Error
	return ruas, err
}

func (r *ruaRepository) FindByID(ctx context.Context, id uint) (*models.Rua, error) {
	var rua models.Rua
	err := r.db.WithContext(ctx).First(&rua, id).Error
	if err != nil {
		return nil, err
	}
	return &rua, nil
}

func (r *ruaRepository) Create(ctx context.Context, rua *models.Rua) error {
	return r.db.WithContext(ctx).Create(rua).Error
}

func (r *ruaRepository) Update(ctx context.Context, rua *models.Rua) error {
	return r.db.WithContext(ctx).Save(rua).Error
}

func (r *ruaRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Rua{}, id).Error
}

// ContarDistritosDistintos conta quantos distritos diferentes existem de fato
// na tabela ruas (em vez de um número fixo no código).
func (r *ruaRepository) ContarDistritosDistintos(ctx context.Context) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&models.Rua{}).
		Distinct("distrito").
		Where("distrito IS NOT NULL AND distrito <> ''").
		Count(&total).Error
	return total, err
}