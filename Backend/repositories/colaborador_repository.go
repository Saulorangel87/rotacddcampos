package repositories

import (
	"context"
	"strings"

	"github.com/empresa/rotas-entrega/models"
	"gorm.io/gorm"
)

type ColaboradorRepository interface {
	FindAll(ctx context.Context, filters map[string]string) ([]models.Colaborador, error)
	FindByID(ctx context.Context, id uint) (*models.Colaborador, error)
	FindAniversariantes(ctx context.Context, mes, dia int) ([]models.Colaborador, error)
}

type colaboradorRepository struct {
	db *gorm.DB
}

func NewColaboradorRepository(db *gorm.DB) ColaboradorRepository {
	return &colaboradorRepository{db: db}
}

func (r *colaboradorRepository) FindAll(ctx context.Context, filters map[string]string) ([]models.Colaborador, error) {
	var colaboradores []models.Colaborador
	query := r.db.WithContext(ctx).Model(&models.Colaborador{})

	if nome, ok := filters["nome"]; ok && nome != "" {
		query = query.Where("nome ILIKE ?", "%"+nome+"%")
	}
	if matricula, ok := filters["matricula"]; ok && matricula != "" {
		query = query.Where("matricula LIKE ?", "%"+matricula+"%")
	}
	if carteiro, ok := filters["carteiro"]; ok && strings.ToLower(carteiro) == "true" {
		query = query.Where("funcao ILIKE ? OR funcao ILIKE ?", "%motorizado%", "%ciclista%")
	}

	err := query.Order("nome asc").Find(&colaboradores).Error
	return colaboradores, err
}

func (r *colaboradorRepository) FindByID(ctx context.Context, id uint) (*models.Colaborador, error) {
	var colaborador models.Colaborador
	err := r.db.WithContext(ctx).First(&colaborador, id).Error
	if err != nil {
		return nil, err
	}
	return &colaborador, nil
}

// FindAniversariantes retorna colaboradores cuja data_nascimento cai no mês/dia informados,
// ignorando o ano (é assim que aniversário funciona). Usa EXTRACT direto no Postgres.
func (r *colaboradorRepository) FindAniversariantes(ctx context.Context, mes, dia int) ([]models.Colaborador, error) {
	var colaboradores []models.Colaborador
	err := r.db.WithContext(ctx).
		Where("data_nascimento IS NOT NULL").
		Where("EXTRACT(MONTH FROM data_nascimento) = ?", mes).
		Where("EXTRACT(DAY FROM data_nascimento) = ?", dia).
		Order("nome asc").
		Find(&colaboradores).Error
	return colaboradores, err
}
