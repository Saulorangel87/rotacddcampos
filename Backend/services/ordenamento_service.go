package services

import (
	"context"
	"errors"

	"github.com/empresa/rotas-entrega/models"
	"github.com/empresa/rotas-entrega/repositories"
)

var ErrOrdenamentoJaExiste = errors.New("já existe um ordenamento em andamento")

type OrdenamentoService interface {
	GetAtivo(ctx context.Context, usuarioID uint) (*models.Ordenamento, error)
	Criar(ctx context.Context, usuarioID uint) (*models.Ordenamento, error)
}

type ordenamentoService struct {
	repo repositories.OrdenamentoRepository
}

func NewOrdenamentoService(repo repositories.OrdenamentoRepository) OrdenamentoService {
	return &ordenamentoService{repo: repo}
}

func (s *ordenamentoService) GetAtivo(ctx context.Context, usuarioID uint) (*models.Ordenamento, error) {
	return s.repo.FindAtivoByUsuario(ctx, usuarioID)
}

func (s *ordenamentoService) Criar(ctx context.Context, usuarioID uint) (*models.Ordenamento, error) {
	existente, err := s.repo.FindAtivoByUsuario(ctx, usuarioID)
	if err != nil {
		return nil, err
	}
	if existente != nil {
		return nil, ErrOrdenamentoJaExiste
	}

	ordenamento := &models.Ordenamento{
		UsuarioID: usuarioID,
		Status:    models.StatusOrdenamentoEmAndamento,
	}
	if err := s.repo.Create(ctx, ordenamento); err != nil {
		return nil, err
	}
	return ordenamento, nil
}
