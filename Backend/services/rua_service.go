package services

import (
	"context"
	"errors"
	"strings"

	"github.com/empresa/rotas-entrega/models"
	"github.com/empresa/rotas-entrega/repositories"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type RuaService interface {
	List(ctx context.Context, nome, cep, distrito string) ([]models.Rua, error)
	GetByID(ctx context.Context, id uint) (*models.Rua, error)
	Create(ctx context.Context, dto CreateRuaDTO) (*models.Rua, error)
	Update(ctx context.Context, id uint, dto UpdateRuaDTO) (*models.Rua, error)
	Delete(ctx context.Context, id uint) error
}

type ruaService struct {
	repo          repositories.RuaRepository
	historicoRepo repositories.HistoricoRepository
}

func NewRuaService(repo repositories.RuaRepository, historicoRepo repositories.HistoricoRepository) RuaService {
	return &ruaService{repo: repo, historicoRepo: historicoRepo}
}

type CreateRuaDTO struct {
	NomeRua    string `json:"nome_rua" validate:"required,max=255"`
	Bairro     string `json:"bairro" validate:"max=100"`
	CEP        string `json:"cep" validate:"required,max=20"`
	Distrito   string `json:"distrito" validate:"required,max=100"`
	Rota       string `json:"rota" validate:"max=50"`
	Observacao string `json:"observacao"`
}

type UpdateRuaDTO struct {
	NomeRua    string `json:"nome_rua" validate:"omitempty,max=255"`
	Bairro     string `json:"bairro" validate:"max=100"`
	CEP        string `json:"cep" validate:"omitempty,max=20"`
	Distrito   string `json:"distrito" validate:"omitempty,max=100"`
	Rota       string `json:"rota" validate:"max=50"`
	Observacao string `json:"observacao"`
	Usuario    string `json:"usuario" validate:"max=150"`
}

func (s *ruaService) List(ctx context.Context, nome, cep, distrito string) ([]models.Rua, error) {
	filters := map[string]string{
		"nome":     strings.TrimSpace(nome),
		"cep":      strings.TrimSpace(cep),
		"distrito": strings.TrimSpace(distrito),
	}
	return s.repo.FindAll(ctx, filters)
}

func (s *ruaService) GetByID(ctx context.Context, id uint) (*models.Rua, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ruaService) Create(ctx context.Context, dto CreateRuaDTO) (*models.Rua, error) {
	if err := validate.Struct(dto); err != nil {
		return nil, err
	}

	rua := &models.Rua{
		NomeRua:    strings.TrimSpace(dto.NomeRua),
		Bairro:     strings.TrimSpace(dto.Bairro),
		CEP:        strings.TrimSpace(dto.CEP),
		Distrito:   strings.TrimSpace(dto.Distrito),
		Rota:       strings.TrimSpace(dto.Rota),
		Observacao: strings.TrimSpace(dto.Observacao),
	}

	if err := s.repo.Create(ctx, rua); err != nil {
		return nil, err
	}
	return rua, nil
}

func (s *ruaService) Update(ctx context.Context, id uint, dto UpdateRuaDTO) (*models.Rua, error) {
	if err := validate.Struct(dto); err != nil {
		return nil, err
	}

	rua, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("rua não encontrada")
	}

	distritoAntigo := rua.Distrito

	if dto.NomeRua != "" {
		rua.NomeRua = strings.TrimSpace(dto.NomeRua)
	}
	if dto.Bairro != "" {
		rua.Bairro = strings.TrimSpace(dto.Bairro)
	}
	if dto.CEP != "" {
		rua.CEP = strings.TrimSpace(dto.CEP)
	}
	if dto.Distrito != "" {
		rua.Distrito = strings.TrimSpace(dto.Distrito)
	}
	if dto.Rota != "" {
		rua.Rota = strings.TrimSpace(dto.Rota)
	}
	rua.Observacao = strings.TrimSpace(dto.Observacao)

	if err := s.repo.Update(ctx, rua); err != nil {
		return nil, err
	}

	// Se o distrito realmente mudou, registra no histórico — best-effort:
	// se isso falhar, não desfaz a atualização da rua (que já foi salva).
	if distritoAntigo != "" && rua.Distrito != "" && distritoAntigo != rua.Distrito && s.historicoRepo != nil {
		usuario := strings.TrimSpace(dto.Usuario)
		if usuario == "" {
			usuario = "Não identificado (sem login)"
		}
		_ = s.historicoRepo.Create(ctx, &models.HistoricoAlteracao{
			Tipo:            "rua",
			RuaID:           rua.ID,
			NomeRua:         rua.NomeRua,
			DistritoOrigem:  distritoAntigo,
			DistritoDestino: rua.Distrito,
			Usuario:         usuario,
		})
	}

	return rua, nil
}

func (s *ruaService) Delete(ctx context.Context, id uint) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("rua não encontrada")
	}
	return s.repo.Delete(ctx, id)
}
