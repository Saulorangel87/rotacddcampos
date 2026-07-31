package services

import (
	"context"

	"github.com/empresa/rotas-entrega/models"
	"github.com/empresa/rotas-entrega/repositories"
)

type HistoricoService interface {
	List(ctx context.Context, pagina, limite int) ([]models.HistoricoAlteracao, int64, error)
}

type historicoService struct {
	repo repositories.HistoricoRepository
}

func NewHistoricoService(repo repositories.HistoricoRepository) HistoricoService {
	return &historicoService{repo: repo}
}

func (s *historicoService) List(ctx context.Context, pagina, limite int) ([]models.HistoricoAlteracao, int64, error) {
	return s.repo.FindAll(ctx, pagina, limite)
}
