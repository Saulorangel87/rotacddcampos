package services

import (
	"context"
	"errors"
	"strings"

	"github.com/empresa/rotas-entrega/models"
	"github.com/empresa/rotas-entrega/repositories"
)

var categoriasValidas = map[models.CategoriaObservacao]bool{
	models.ObsAcesso:          true,
	models.ObsSeguranca:       true,
	models.ObsNumeroIrregular: true,
	models.ObsVariosNomes:     true,
	models.ObsOutros:          true,
}

type NovaObservacaoDTO struct {
	Categoria string `json:"categoria"`
	Texto     string `json:"texto"`
}

type RuaObservacaoService interface {
	Listar(ctx context.Context, ruaID uint) ([]models.RuaObservacao, error)
	Adicionar(ctx context.Context, ruaID uint, dto NovaObservacaoDTO, criadoPor string) (*models.RuaObservacao, error)
	Excluir(ctx context.Context, id uint) error
}

type ruaObservacaoService struct {
	repo    repositories.RuaObservacaoRepository
	ruaRepo repositories.RuaRepository
}

func NewRuaObservacaoService(repo repositories.RuaObservacaoRepository, ruaRepo repositories.RuaRepository) RuaObservacaoService {
	return &ruaObservacaoService{repo: repo, ruaRepo: ruaRepo}
}

func (s *ruaObservacaoService) Listar(ctx context.Context, ruaID uint) ([]models.RuaObservacao, error) {
	return s.repo.FindByRuaID(ctx, ruaID)
}

func (s *ruaObservacaoService) Adicionar(ctx context.Context, ruaID uint, dto NovaObservacaoDTO, criadoPor string) (*models.RuaObservacao, error) {
	if _, err := s.ruaRepo.FindByID(ctx, ruaID); err != nil {
		return nil, errors.New("rua não encontrada")
	}

	categoria := models.CategoriaObservacao(strings.ToLower(strings.TrimSpace(dto.Categoria)))
	if !categoriasValidas[categoria] {
		return nil, errors.New("categoria inválida — use acesso, seguranca, numero_irregular, varios_nomes ou outros")
	}

	texto := strings.TrimSpace(dto.Texto)
	if texto == "" {
		return nil, errors.New("texto da observação é obrigatório")
	}

	obs := &models.RuaObservacao{
		RuaID:     ruaID,
		Categoria: categoria,
		Texto:     texto,
		CriadoPor: criadoPor,
	}
	if err := s.repo.Create(ctx, obs); err != nil {
		return nil, err
	}
	return obs, nil
}

func (s *ruaObservacaoService) Excluir(ctx context.Context, id uint) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return errors.New("observação não encontrada")
	}
	return s.repo.Delete(ctx, id)
}
