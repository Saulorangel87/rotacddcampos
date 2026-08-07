package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/empresa/rotas-entrega/models"
	"github.com/empresa/rotas-entrega/repositories"
)

type NovoLancamentoDTO struct {
	Matricula      string `json:"matricula"`
	Tipo           string `json:"tipo"` // "credito" ou "debito"
	Quantidade     int    `json:"quantidade"`
	Motivo         string `json:"motivo"`
	DataReferencia string `json:"data_referencia"` // DD/MM/AAAA, opcional
}

// SaldoFolgas é a resposta da consulta pública por matrícula: nome (só pra
// confirmar visualmente que achou a pessoa certa), saldo atual já calculado
// e o extrato completo que sustenta esse número.
type SaldoFolgas struct {
	Matricula   string                   `json:"matricula"`
	Nome        string                   `json:"nome"`
	Saldo       int                      `json:"saldo"`
	Lancamentos []models.FolgaLancamento `json:"lancamentos"`
}

type FolgaService interface {
	ConsultarSaldo(ctx context.Context, matricula string) (*SaldoFolgas, error)
	Lancar(ctx context.Context, dto NovoLancamentoDTO, criadoPor string) (*models.FolgaLancamento, error)
	Excluir(ctx context.Context, id uint, excluidoPor string) error
}

type folgaService struct {
	repo            repositories.FolgaRepository
	colaboradorRepo repositories.ColaboradorRepository
	historicoRepo   repositories.HistoricoRepository
}

func NewFolgaService(repo repositories.FolgaRepository, colaboradorRepo repositories.ColaboradorRepository, historicoRepo repositories.HistoricoRepository) FolgaService {
	return &folgaService{repo: repo, colaboradorRepo: colaboradorRepo, historicoRepo: historicoRepo}
}

// ConsultarSaldo exige a matrícula exata — não é uma busca parcial tipo
// "listar quem bate com esse pedaço", de propósito: essa rota é pública e não
// deve virar um jeito de listar todo mundo sem saber a matrícula de ninguém.
func (s *folgaService) ConsultarSaldo(ctx context.Context, matricula string) (*SaldoFolgas, error) {
	matricula = strings.TrimSpace(matricula)
	if matricula == "" {
		return nil, errors.New("matrícula é obrigatória")
	}

	colaboradores, err := s.colaboradorRepo.FindAll(ctx, map[string]string{"matricula": matricula})
	if err != nil {
		return nil, err
	}
	nome := ""
	for _, c := range colaboradores {
		if c.Matricula == matricula {
			nome = c.Nome
			break
		}
	}
	if nome == "" {
		return nil, errors.New("matrícula não encontrada")
	}

	lancamentos, err := s.repo.FindByMatricula(ctx, matricula)
	if err != nil {
		return nil, err
	}

	saldo := 0
	for _, l := range lancamentos {
		if l.Tipo == models.FolgaCredito {
			saldo += l.Quantidade
		} else {
			saldo -= l.Quantidade
		}
	}

	return &SaldoFolgas{
		Matricula:   matricula,
		Nome:        nome,
		Saldo:       saldo,
		Lancamentos: lancamentos,
	}, nil
}

func (s *folgaService) Lancar(ctx context.Context, dto NovoLancamentoDTO, criadoPor string) (*models.FolgaLancamento, error) {
	matricula := strings.TrimSpace(dto.Matricula)
	motivo := strings.TrimSpace(dto.Motivo)
	tipo := models.TipoFolga(strings.ToLower(strings.TrimSpace(dto.Tipo)))

	if matricula == "" {
		return nil, errors.New("matrícula é obrigatória")
	}
	if tipo != models.FolgaCredito && tipo != models.FolgaDebito {
		return nil, errors.New("tipo deve ser 'credito' ou 'debito'")
	}
	if motivo == "" {
		return nil, errors.New("motivo/justificativa é obrigatório")
	}
	if dto.Quantidade <= 0 {
		return nil, errors.New("quantidade deve ser maior que zero")
	}

	lancamento := &models.FolgaLancamento{
		Matricula:  matricula,
		Tipo:       tipo,
		Quantidade: dto.Quantidade,
		Motivo:     motivo,
		CriadoPor:  criadoPor,
	}

	if dto.DataReferencia != "" {
		data, err := time.Parse("02/01/2006", dto.DataReferencia)
		if err != nil {
			return nil, errors.New("data de referência inválida, use DD/MM/AAAA")
		}
		lancamento.DataReferencia = &data
	}

	if err := s.repo.Create(ctx, lancamento); err != nil {
		return nil, err
	}

	// Toda inclusão de folga vai pro mesmo histórico auditável das ruas —
	// se alguém lançar folga indevidamente, fica registrado quem foi e quando,
	// mesmo que a pessoa apague o lançamento depois (a exclusão também é logada).
	sinal := "+"
	if tipo == models.FolgaDebito {
		sinal = "-"
	}
	s.registrarHistorico(ctx, fmt.Sprintf(
		"Lançou %s%d folga(s) para matrícula %s — %s",
		sinal, lancamento.Quantidade, matricula, motivo,
	), criadoPor)

	return lancamento, nil
}

func (s *folgaService) Excluir(ctx context.Context, id uint, excluidoPor string) error {
	lancamento, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("lançamento não encontrado")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	s.registrarHistorico(ctx, fmt.Sprintf(
		"Excluiu lançamento de %s (%d) da matrícula %s — %s",
		lancamento.Tipo, lancamento.Quantidade, lancamento.Matricula, lancamento.Motivo,
	), excluidoPor)

	return nil
}

// registrarHistorico nunca falha a operação principal por causa de um erro
// no log — se o lançamento de folga já foi salvo, ele fica salvo mesmo que a
// gravação do histórico falhe por algum motivo raro.
func (s *folgaService) registrarHistorico(ctx context.Context, descricao, usuario string) {
	if s.historicoRepo == nil {
		return
	}
	_ = s.historicoRepo.Create(ctx, &models.HistoricoAlteracao{
		Tipo:      "folga",
		Descricao: descricao,
		Usuario:   usuario,
	})
}
