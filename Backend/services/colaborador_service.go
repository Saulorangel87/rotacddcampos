package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/empresa/rotas-entrega/models"
	"github.com/empresa/rotas-entrega/repositories"
)

type NovoColaboradorDTO struct {
	Nome           string `json:"nome"`
	Matricula      string `json:"matricula"`
	Funcao         string `json:"funcao"`
	Cargo          string `json:"cargo"`
	DataAdmissao   string `json:"data_admissao"`   // formato DD/MM/AAAA, opcional
	DataNascimento string `json:"data_nascimento"` // formato DD/MM/AAAA, opcional
}

type ColaboradorService interface {
	List(ctx context.Context, nome, matricula, carteiro string) ([]models.Colaborador, error)
	GetByID(ctx context.Context, id uint) (*models.Colaborador, error)
	AniversariantesDeHoje(ctx context.Context) ([]models.Colaborador, error)
	AniversariantesDaData(ctx context.Context, mes, dia int) ([]models.Colaborador, error)
	Create(ctx context.Context, dto NovoColaboradorDTO) (*models.Colaborador, error)
	Delete(ctx context.Context, id uint) error
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

func (s *colaboradorService) Create(ctx context.Context, dto NovoColaboradorDTO) (*models.Colaborador, error) {
	nome := strings.TrimSpace(dto.Nome)
	matricula := strings.TrimSpace(dto.Matricula)
	if nome == "" || matricula == "" {
		return nil, errors.New("nome e matrícula são obrigatórios")
	}

	colaborador := &models.Colaborador{
		Nome:      nome,
		Matricula: matricula,
		Funcao:    strings.TrimSpace(dto.Funcao),
		Cargo:     strings.TrimSpace(dto.Cargo),
	}

	if dto.DataAdmissao != "" {
		data, err := time.Parse("02/01/2006", dto.DataAdmissao)
		if err != nil {
			return nil, errors.New("data de admissão inválida, use DD/MM/AAAA")
		}
		colaborador.DataAdmissao = &data
	}

	if dto.DataNascimento != "" {
		data, err := time.Parse("02/01/2006", dto.DataNascimento)
		if err != nil {
			return nil, errors.New("data de nascimento inválida, use DD/MM/AAAA")
		}
		colaborador.DataNascimento = &data
	}

	if err := s.repo.Create(ctx, colaborador); err != nil {
		return nil, err
	}
	return colaborador, nil
}

func (s *colaboradorService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}
