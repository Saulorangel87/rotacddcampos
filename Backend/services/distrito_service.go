package services

import (
	"context"

	"github.com/empresa/rotas-entrega/models"
	"github.com/empresa/rotas-entrega/repositories"
)

type DistritoService interface {
	List(ctx context.Context) ([]models.Distrito, error)
	GetByCodigo(ctx context.Context, codigo string) (*models.Distrito, error)
}

type distritoService struct {
	repo repositories.DistritoRepository
}

func NewDistritoService(repo repositories.DistritoRepository) DistritoService {
	return &distritoService{repo: repo}
}

func (s *distritoService) List(ctx context.Context) ([]models.Distrito, error) {
	return s.repo.FindAll(ctx)
}

func (s *distritoService) GetByCodigo(ctx context.Context, codigo string) (*models.Distrito, error) {
	return s.repo.FindByCodigo(ctx, codigo)
}
