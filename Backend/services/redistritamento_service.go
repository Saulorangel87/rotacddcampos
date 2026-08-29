package services

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/empresa/rotas-entrega/models"
	"github.com/empresa/rotas-entrega/repositories"
)

var (
	ErrPlanoJaExiste       = errors.New("já existe um plano de redistritamento em andamento")
	ErrAlvoInvalido        = errors.New("a quantidade alvo precisa ser menor que a quantidade atual de distritos ativos")
	ErrPlanoNaoEncontrado  = errors.New("plano de redistritamento não encontrado")
	ErrRuasSemDestino      = errors.New("ainda existem ruas órfãs sem distrito de destino definido")
	ErrDestinoInvalido     = errors.New("distrito de destino inválido: precisa ser um distrito ativo que não está sendo extinto")
	ErrPlanoJaAplicado     = errors.New("este plano já foi aplicado e não pode mais ser alterado")
	ErrPlanoNaoConcluido   = errors.New("o plano precisa ser concluído antes de aplicar — finalize a realocação de todas as ruas primeiro")
)

type RedistritamentoService interface {
	GetPlanoAtivo(ctx context.Context) (*models.PlanoRedistritamento, error)
	CriarPlano(ctx context.Context, quantidadeAlvo int, criadoPor string) (*models.PlanoRedistritamento, error)
	ReatribuirRua(ctx context.Context, planoID, planoRuaID uint, distritoDestino string) error
	Concluir(ctx context.Context, planoID uint) (*models.PlanoRedistritamento, error)
	Aplicar(ctx context.Context, planoID uint, matricula string) error
	// CancelarPlano descarta um rascunho/concluído e volta pra tela inicial —
	// seguro porque nada real muda no banco antes de Aplicar.
	CancelarPlano(ctx context.Context, planoID uint) error
}

type redistritamentoService struct {
	repo         repositories.RedistritamentoRepository
	distritoRepo repositories.DistritoRepository
	ruaRepo      repositories.RuaRepository
}

func NewRedistritamentoService(repo repositories.RedistritamentoRepository, distritoRepo repositories.DistritoRepository, ruaRepo repositories.RuaRepository) RedistritamentoService {
	return &redistritamentoService{repo: repo, distritoRepo: distritoRepo, ruaRepo: ruaRepo}
}

func (s *redistritamentoService) GetPlanoAtivo(ctx context.Context) (*models.PlanoRedistritamento, error) {
	return s.repo.FindPlanoAtivo(ctx)
}

func (s *redistritamentoService) CriarPlano(ctx context.Context, quantidadeAlvo int, criadoPor string) (*models.PlanoRedistritamento, error) {
	existente, err := s.repo.FindPlanoAtivo(ctx)
	if err != nil {
		return nil, err
	}
	if existente != nil {
		return nil, ErrPlanoJaExiste
	}

	todos, err := s.distritoRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	var ativosCodigos []string
	for _, d := range todos {
		if d.Ativo {
			ativosCodigos = append(ativosCodigos, d.Codigo)
		}
	}

	quantidadeAtual := len(ativosCodigos)
	if quantidadeAlvo <= 0 || quantidadeAlvo >= quantidadeAtual {
		return nil, ErrAlvoInvalido
	}

	// Corte global: os N códigos de maior valor entre os ativos são extintos.
	sort.Sort(sort.Reverse(sort.StringSlice(ativosCodigos)))
	numExtinguir := quantidadeAtual - quantidadeAlvo
	extintos := ativosCodigos[:numExtinguir]

	ruasOrfas, err := s.ruaRepo.FindByDistritosExato(ctx, extintos)
	if err != nil {
		return nil, err
	}

	plano := &models.PlanoRedistritamento{
		Tipo:              "reducao",
		Status:            models.StatusRedistritamentoRascunho,
		QuantidadeAtual:   quantidadeAtual,
		QuantidadeAlvo:    quantidadeAlvo,
		DistritosExtintos: strings.Join(extintos, ","),
		CriadoPor:         criadoPor,
	}

	for _, rua := range ruasOrfas {
		plano.Ruas = append(plano.Ruas, models.PlanoRedistritamentoRua{
			RuaID:          rua.ID,
			NomeRua:        rua.NomeRua,
			Bairro:         rua.Bairro,
			DistritoOrigem: rua.Distrito,
		})
	}

	if err := s.repo.CreatePlano(ctx, plano); err != nil {
		return nil, err
	}
	return plano, nil
}

func (s *redistritamentoService) ReatribuirRua(ctx context.Context, planoID, planoRuaID uint, distritoDestino string) error {
	plano, err := s.repo.FindPlanoByID(ctx, planoID)
	if err != nil {
		return ErrPlanoNaoEncontrado
	}
	if plano.Status == models.StatusRedistritamentoAplicado {
		return ErrPlanoJaAplicado
	}

	extintos := strings.Split(plano.DistritosExtintos, ",")
	for _, ext := range extintos {
		if ext == distritoDestino {
			return ErrDestinoInvalido
		}
	}

	distrito, err := s.distritoRepo.FindByCodigo(ctx, distritoDestino)
	if err != nil || distrito == nil || !distrito.Ativo {
		return ErrDestinoInvalido
	}

	return s.repo.UpdateDestinoRua(ctx, planoRuaID, distritoDestino)
}

func (s *redistritamentoService) Concluir(ctx context.Context, planoID uint) (*models.PlanoRedistritamento, error) {
	plano, err := s.repo.FindPlanoByID(ctx, planoID)
	if err != nil {
		return nil, ErrPlanoNaoEncontrado
	}
	if plano.Status == models.StatusRedistritamentoAplicado {
		return nil, ErrPlanoJaAplicado
	}

	semDestino, err := s.repo.CountRuasSemDestino(ctx, planoID)
	if err != nil {
		return nil, err
	}
	if semDestino > 0 {
		return nil, ErrRuasSemDestino
	}

	agora := time.Now()
	plano.Status = models.StatusRedistritamentoConcluido
	plano.ConcluidoEm = &agora
	if err := s.repo.UpdatePlano(ctx, plano); err != nil {
		return nil, err
	}
	return plano, nil
}

func (s *redistritamentoService) Aplicar(ctx context.Context, planoID uint, matricula string) error {
	plano, err := s.repo.FindPlanoByID(ctx, planoID)
	if err != nil {
		return ErrPlanoNaoEncontrado
	}
	if plano.Status == models.StatusRedistritamentoAplicado {
		return ErrPlanoJaAplicado
	}
	if plano.Status != models.StatusRedistritamentoConcluido {
		return ErrPlanoNaoConcluido
	}

	semDestino, err := s.repo.CountRuasSemDestino(ctx, planoID)
	if err != nil {
		return err
	}
	if semDestino > 0 {
		return ErrRuasSemDestino
	}

	agora := time.Now()
	plano.AplicadoEm = &agora
	return s.repo.Aplicar(ctx, plano, matricula)
}

func (s *redistritamentoService) CancelarPlano(ctx context.Context, planoID uint) error {
	plano, err := s.repo.FindPlanoByID(ctx, planoID)
	if err != nil {
		return ErrPlanoNaoEncontrado
	}
	if plano.Status == models.StatusRedistritamentoAplicado {
		return ErrPlanoJaAplicado
	}
	return s.repo.DeletePlano(ctx, planoID)
}
