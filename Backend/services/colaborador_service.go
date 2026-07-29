package services

import (
	"context"
	"strings"
	"time"

	"github.com/empresa/rotas-entrega/models"
	"github.com/empresa/rotas-entrega/repositories"
)

type ColaboradorService interface {
	List(ctx context.Context, nome, matricula, carteiro string) ([]models.Colaborador, error)
	GetByID(ctx context.Context, id uint) (*models.Colaborador, error)
	AniversariantesDeHoje(ctx context.Context) ([]models.Colaborador, error)
	AniversariantesDaData(ctx context.Context, mes, dia int) ([]models.Colaborador, error)
}

type colaboradorService struct {
	repo repositories.ColaboradorRepository
}

func NewColaboradorService(repo repositories.ColaboradorRepository) ColaboradorService {
	return &colaboradorService{repo: repo}
}

func (s *colaboradorService) List(ctx context.Context, nome, matricula, carteiro string) ([]models.Colaborador, error) {
	filters := map[string]string{
		"nome":      strings.TrimSpace(nome),
		"matricula": strings.TrimSpace(matricula),
		"carteiro":  strings.TrimSpace(carteiro),
	}
	return s.repo.FindAll(ctx, filters)
}

func (s *colaboradorService) GetByID(ctx context.Context, id uint) (*models.Colaborador, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *colaboradorService) AniversariantesDeHoje(ctx context.Context) ([]models.Colaborador, error) {
	hoje := time.Now()
	return s.repo.FindAniversariantes(ctx, int(hoje.Month()), hoje.Day())
}

func (s *colaboradorService) AniversariantesDaData(ctx context.Context, mes, dia int) ([]models.Colaborador, error) {
	return s.repo.FindAniversariantes(ctx, mes, dia)
}
