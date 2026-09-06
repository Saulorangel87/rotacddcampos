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
	ErrOrdenamentoJaExiste      = errors.New("já existe um ordenamento em andamento")
	ErrOrdenamentoNaoEncontrado = errors.New("ordenamento não encontrado")
	ErrObjetoNaoEncontrado      = errors.New("encomenda não encontrada")
	ErrEntradaVazia             = errors.New("informe o endereço da encomenda")
)

type AdicionarObjetoDTO struct {
	Entrada string `json:"entrada"`
	Origem  string `json:"origem"`
}

type RuaAgrupada struct {
	Chave      string `json:"chave"`
	NomeRua    string `json:"nome_rua"`
	Quantidade int    `json:"quantidade"`
}

type OrdenamentoDetalhe struct {
	ID             uint                       `json:"id"`
	Status         string                     `json:"status"`
	CreatedAt      time.Time                  `json:"created_at"`
	UpdatedAt      time.Time                  `json:"updated_at"`
	TotalObjetos   int                        `json:"total_objetos"`
	TotalRuas      int                        `json:"total_ruas"`
	TotalPendentes int                        `json:"total_pendentes"`
	Objetos        []models.ObjetoOrdenamento `json:"objetos"`
	Ruas           []RuaAgrupada              `json:"ruas"`
}

type OrdenamentoService interface {
	GetAtivo(ctx context.Context, usuarioID uint) (*OrdenamentoDetalhe, error)
	Criar(ctx context.Context, usuarioID uint) (*OrdenamentoDetalhe, error)
	AdicionarObjeto(ctx context.Context, usuarioID, ordenamentoID uint, dto AdicionarObjetoDTO) (*OrdenamentoDetalhe, error)
	ExcluirObjeto(ctx context.Context, usuarioID, ordenamentoID, objetoID uint) (*OrdenamentoDetalhe, error)
}

type ordenamentoService struct {
	repo     repositories.OrdenamentoRepository
	resolver EnderecoResolver
}

func NewOrdenamentoService(repo repositories.OrdenamentoRepository, resolver EnderecoResolver) OrdenamentoService {
	return &ordenamentoService{repo: repo, resolver: resolver}
}

func (s *ordenamentoService) GetAtivo(ctx context.Context, usuarioID uint) (*OrdenamentoDetalhe, error) {
	ordenamento, err := s.repo.FindAtivoByUsuario(ctx, usuarioID)
	if err != nil || ordenamento == nil {
		return nil, err
	}
	return s.montarDetalhe(ctx, ordenamento)
}

func (s *ordenamentoService) Criar(ctx context.Context, usuarioID uint) (*OrdenamentoDetalhe, error) {
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
	return s.montarDetalhe(ctx, ordenamento)
}

func (s *ordenamentoService) AdicionarObjeto(ctx context.Context, usuarioID, ordenamentoID uint, dto AdicionarObjetoDTO) (*OrdenamentoDetalhe, error) {
	entrada := strings.TrimSpace(dto.Entrada)
	if entrada == "" {
		return nil, ErrEntradaVazia
	}
	ordenamento, err := s.validarOrdenamentoAtivo(ctx, usuarioID, ordenamentoID)
	if err != nil {
		return nil, err
	}

	origem := strings.TrimSpace(dto.Origem)
	if origem != models.OrigemEntradaVoz && origem != models.OrigemEntradaScanner {
		origem = models.OrigemEntradaManual
	}
	resolucao, err := s.resolver.Resolver(ctx, entrada)
	if err != nil {
		return nil, err
	}
	objeto := &models.ObjetoOrdenamento{
		OrdenamentoID:    ordenamento.ID,
		RuaID:            resolucao.RuaID,
		TextoEntrada:     entrada,
		NomeRua:          resolucao.NomeRua,
		ChaveAgrupamento: resolucao.ChaveAgrupamento,
		Numero:           resolucao.Numero,
		CEP:              resolucao.CEP,
		OrigemEntrada:    origem,
		StatusResolucao:  resolucao.Status,
		MotivoPendencia:  resolucao.MotivoPendencia,
	}
	if err := s.repo.CreateObjeto(ctx, objeto); err != nil {
		return nil, err
	}
	return s.montarDetalhe(ctx, ordenamento)
}

func (s *ordenamentoService) ExcluirObjeto(ctx context.Context, usuarioID, ordenamentoID, objetoID uint) (*OrdenamentoDetalhe, error) {
	ordenamento, err := s.validarOrdenamentoAtivo(ctx, usuarioID, ordenamentoID)
	if err != nil {
		return nil, err
	}
	excluido, err := s.repo.DeleteObjeto(ctx, ordenamentoID, objetoID)
	if err != nil {
		return nil, err
	}
	if !excluido {
		return nil, ErrObjetoNaoEncontrado
	}
	return s.montarDetalhe(ctx, ordenamento)
}

func (s *ordenamentoService) validarOrdenamentoAtivo(ctx context.Context, usuarioID, ordenamentoID uint) (*models.Ordenamento, error) {
	ordenamento, err := s.repo.FindAtivoByUsuario(ctx, usuarioID)
	if err != nil {
		return nil, err
	}
	if ordenamento == nil || ordenamento.ID != ordenamentoID {
		return nil, ErrOrdenamentoNaoEncontrado
	}
	return ordenamento, nil
}

func (s *ordenamentoService) montarDetalhe(ctx context.Context, ordenamento *models.Ordenamento) (*OrdenamentoDetalhe, error) {
	objetos, err := s.repo.ListObjetos(ctx, ordenamento.ID)
	if err != nil {
		return nil, err
	}
	agrupadas := make(map[string]*RuaAgrupada)
	pendentes := 0
	for _, objeto := range objetos {
		if objeto.StatusResolucao != models.StatusResolucaoIdentificado || objeto.ChaveAgrupamento == "" {
			pendentes++
			continue
		}
		rua := agrupadas[objeto.ChaveAgrupamento]
		if rua == nil {
			rua = &RuaAgrupada{Chave: objeto.ChaveAgrupamento, NomeRua: objeto.NomeRua}
			agrupadas[objeto.ChaveAgrupamento] = rua
		}
		rua.Quantidade++
	}
	ruas := make([]RuaAgrupada, 0, len(agrupadas))
	for _, rua := range agrupadas {
		ruas = append(ruas, *rua)
	}
	sort.Slice(ruas, func(i, j int) bool { return ruas[i].NomeRua < ruas[j].NomeRua })

	return &OrdenamentoDetalhe{
		ID:             ordenamento.ID,
		Status:         ordenamento.Status,
		CreatedAt:      ordenamento.CreatedAt,
		UpdatedAt:      ordenamento.UpdatedAt,
		TotalObjetos:   len(objetos),
		TotalRuas:      len(ruas),
		TotalPendentes: pendentes,
		Objetos:        objetos,
		Ruas:           ruas,
	}, nil
}
