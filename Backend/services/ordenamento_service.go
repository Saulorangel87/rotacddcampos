package services

import (
	"context"
	"errors"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/empresa/rotas-entrega/models"
	"github.com/empresa/rotas-entrega/repositories"
)

var (
	ErrOrdenamentoJaExiste       = errors.New("já existe um ordenamento em andamento")
	ErrOrdenamentoNaoEncontrado  = errors.New("ordenamento não encontrado")
	ErrObjetoNaoEncontrado       = errors.New("encomenda não encontrada")
	ErrEntradaVazia              = errors.New("informe o endereço da encomenda")
	ErrOrdenamentoSemObjetos     = errors.New("adicione ao menos uma rua identificada antes de gerar o ordenamento")
	ErrOrdenamentoComPendencias  = errors.New("revise as encomendas pendentes antes de gerar o ordenamento")
	ErrOrdenamentoSemCoordenadas = errors.New("há ruas sem coordenada disponível para gerar o ordenamento")
	ErrOpcaoRuaInvalida          = errors.New("a rua escolhida não corresponde à encomenda")
)

type AdicionarObjetoDTO struct {
	Entrada string `json:"entrada"`
	Origem  string `json:"origem"`
}

type RuaAgrupada struct {
	Chave           string   `json:"chave"`
	NomeRua         string   `json:"nome_rua"`
	Quantidade      int      `json:"quantidade"`
	Latitude        *float64 `json:"latitude,omitempty"`
	Longitude       *float64 `json:"longitude,omitempty"`
	FonteCoordenada string   `json:"fonte_coordenada,omitempty"`
}

type OrdenamentoDetalhe struct {
	ID                  uint                       `json:"id"`
	Status              string                     `json:"status"`
	CreatedAt           time.Time                  `json:"created_at"`
	UpdatedAt           time.Time                  `json:"updated_at"`
	TotalObjetos        int                        `json:"total_objetos"`
	TotalRuas           int                        `json:"total_ruas"`
	TotalPendentes      int                        `json:"total_pendentes"`
	TotalSemCoordenadas int                        `json:"total_sem_coordenadas"`
	Objetos             []models.ObjetoOrdenamento `json:"objetos"`
	Ruas                []RuaAgrupada              `json:"ruas"`
	OrdemSugerida       []models.ParadaOrdenamento `json:"ordem_sugerida"`
}

type OrdenamentoService interface {
	GetAtivo(ctx context.Context, usuarioID uint) (*OrdenamentoDetalhe, error)
	Criar(ctx context.Context, usuarioID uint) (*OrdenamentoDetalhe, error)
	AdicionarObjeto(ctx context.Context, usuarioID, ordenamentoID uint, dto AdicionarObjetoDTO) (*OrdenamentoDetalhe, error)
	SelecionarRua(ctx context.Context, usuarioID, ordenamentoID, objetoID, ruaID uint) (*OrdenamentoDetalhe, error)
	ExcluirObjeto(ctx context.Context, usuarioID, ordenamentoID, objetoID uint) (*OrdenamentoDetalhe, error)
	Limpar(ctx context.Context, usuarioID, ordenamentoID uint) (*OrdenamentoDetalhe, error)
	GerarOrdem(ctx context.Context, usuarioID, ordenamentoID uint) (*OrdenamentoDetalhe, error)
}

type ordenamentoService struct {
	repo        repositories.OrdenamentoRepository
	resolver    EnderecoResolver
	coordenadas CoordenadaResolver
	paradas     repositories.ParadaOrdenamentoRepository
	otimizador  OtimizadorRota
}

func NewOrdenamentoService(repo repositories.OrdenamentoRepository, resolver EnderecoResolver, coordenadas CoordenadaResolver, paradas repositories.ParadaOrdenamentoRepository, otimizador OtimizadorRota) OrdenamentoService {
	return &ordenamentoService{repo: repo, resolver: resolver, coordenadas: coordenadas, paradas: paradas, otimizador: otimizador}
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
		OpcoesResolucao:  resolucao.Opcoes,
	}
	if resolucao.Status == models.StatusResolucaoIdentificado && s.coordenadas != nil {
		coordenada, erroCoordenada := s.coordenadas.Resolver(ctx, SolicitacaoCoordenada{
			ChaveAgrupamento: resolucao.ChaveAgrupamento,
			NomeRua:          resolucao.NomeRua,
			RuaID:            resolucao.RuaID,
			PermitirExterno:  true,
		})
		if erroCoordenada != nil {
			slog.Warn("não foi possível obter coordenada da rua identificada", "error", erroCoordenada)
		} else if coordenada != nil {
			objeto.Latitude = &coordenada.Latitude
			objeto.Longitude = &coordenada.Longitude
			objeto.FonteCoordenada = coordenada.Fonte
		}
	}
	if err := s.repo.CreateObjeto(ctx, objeto); err != nil {
		return nil, err
	}
	if err := s.limparOrdemSugerida(ctx, ordenamento.ID); err != nil {
		return nil, err
	}
	return s.montarDetalhe(ctx, ordenamento)
}

func (s *ordenamentoService) SelecionarRua(ctx context.Context, usuarioID, ordenamentoID, objetoID, ruaID uint) (*OrdenamentoDetalhe, error) {
	ordenamento, err := s.validarOrdenamentoAtivo(ctx, usuarioID, ordenamentoID)
	if err != nil {
		return nil, err
	}
	selecionavel, ok := s.resolver.(EnderecoResolverSelecionavel)
	if !ok {
		return nil, ErrOpcaoRuaInvalida
	}
	objetos, err := s.repo.ListObjetos(ctx, ordenamento.ID)
	if err != nil {
		return nil, err
	}
	var objeto *models.ObjetoOrdenamento
	for indice := range objetos {
		if objetos[indice].ID == objetoID {
			objeto = &objetos[indice]
			break
		}
	}
	if objeto == nil {
		return nil, ErrObjetoNaoEncontrado
	}
	resolucao, err := selecionavel.Selecionar(ctx, objeto.TextoEntrada, ruaID)
	if err != nil {
		return nil, err
	}
	if resolucao.Status != models.StatusResolucaoIdentificado || resolucao.RuaID == nil || *resolucao.RuaID != ruaID {
		return nil, ErrOpcaoRuaInvalida
	}
	objeto.RuaID = resolucao.RuaID
	objeto.NomeRua = resolucao.NomeRua
	objeto.ChaveAgrupamento = resolucao.ChaveAgrupamento
	objeto.Numero = resolucao.Numero
	objeto.CEP = resolucao.CEP
	objeto.StatusResolucao = resolucao.Status
	objeto.MotivoPendencia = resolucao.MotivoPendencia
	objeto.OpcoesResolucao = nil
	if s.coordenadas != nil {
		coordenada, erroCoordenada := s.coordenadas.Resolver(ctx, SolicitacaoCoordenada{
			ChaveAgrupamento: resolucao.ChaveAgrupamento,
			NomeRua:          resolucao.NomeRua,
			RuaID:            resolucao.RuaID,
			PermitirExterno:  true,
		})
		if erroCoordenada != nil {
			slog.Warn("não foi possível obter coordenada da rua selecionada", "error", erroCoordenada)
		} else if coordenada != nil {
			objeto.Latitude = &coordenada.Latitude
			objeto.Longitude = &coordenada.Longitude
			objeto.FonteCoordenada = coordenada.Fonte
		}
	}
	if err := s.repo.UpdateObjeto(ctx, objeto); err != nil {
		return nil, err
	}
	if err := s.limparOrdemSugerida(ctx, ordenamento.ID); err != nil {
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
	if err := s.limparOrdemSugerida(ctx, ordenamento.ID); err != nil {
		return nil, err
	}
	return s.montarDetalhe(ctx, ordenamento)
}

// Limpar remove os objetos e a ordem gerada, mantendo o ordenamento aberto
// para que o colaborador possa registrar uma nova lista de encomendas.
func (s *ordenamentoService) Limpar(ctx context.Context, usuarioID, ordenamentoID uint) (*OrdenamentoDetalhe, error) {
	ordenamento, err := s.validarOrdenamentoAtivo(ctx, usuarioID, ordenamentoID)
	if err != nil {
		return nil, err
	}
	if err := s.repo.LimparConteudo(ctx, ordenamento.ID); err != nil {
		return nil, err
	}
	return s.montarDetalhe(ctx, ordenamento)
}

// GerarOrdem cria uma sequência por proximidade entre as ruas identificadas,
// sempre partindo do CDD. A sequência é uma sugestão geográfica local e não
// considera trânsito ou condições da malha viária.
func (s *ordenamentoService) GerarOrdem(ctx context.Context, usuarioID, ordenamentoID uint) (*OrdenamentoDetalhe, error) {
	ordenamento, err := s.validarOrdenamentoAtivo(ctx, usuarioID, ordenamentoID)
	if err != nil {
		return nil, err
	}
	detalhe, err := s.montarDetalhe(ctx, ordenamento)
	if err != nil {
		return nil, err
	}
	if detalhe.TotalPendentes > 0 {
		return nil, ErrOrdenamentoComPendencias
	}
	if detalhe.TotalRuas == 0 {
		return nil, ErrOrdenamentoSemObjetos
	}
	if detalhe.TotalSemCoordenadas > 0 {
		return nil, ErrOrdenamentoSemCoordenadas
	}
	if s.paradas == nil || s.otimizador == nil {
		return nil, errors.New("serviço de ordenamento indisponível")
	}

	entrada := make([]ParadaParaOtimizar, 0, len(detalhe.Ruas))
	for _, rua := range detalhe.Ruas {
		entrada = append(entrada, ParadaParaOtimizar{
			Chave: rua.Chave, NomeRua: rua.NomeRua, Quantidade: rua.Quantidade,
			Latitude: *rua.Latitude, Longitude: *rua.Longitude, FonteCoordenada: rua.FonteCoordenada,
		})
	}
	sugeridas := s.otimizador.Otimizar(PontoPartidaCDD(), entrada)
	paradas := make([]models.ParadaOrdenamento, 0, len(sugeridas))
	for indice, parada := range sugeridas {
		paradas = append(paradas, models.ParadaOrdenamento{
			OrdenamentoID: ordenamento.ID, ChaveAgrupamento: parada.Chave, NomeRua: parada.NomeRua,
			QuantidadeObjetos: parada.Quantidade, Latitude: parada.Latitude, Longitude: parada.Longitude,
			OrdemSugerida: indice + 1,
		})
	}
	if err := s.paradas.ReplaceByOrdenamento(ctx, ordenamento.ID, paradas); err != nil {
		return nil, err
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

func (s *ordenamentoService) limparOrdemSugerida(ctx context.Context, ordenamentoID uint) error {
	if s.paradas == nil {
		return nil
	}
	return s.paradas.DeleteByOrdenamento(ctx, ordenamentoID)
}

func (s *ordenamentoService) montarDetalhe(ctx context.Context, ordenamento *models.Ordenamento) (*OrdenamentoDetalhe, error) {
	objetos, err := s.repo.ListObjetos(ctx, ordenamento.ID)
	if err != nil {
		return nil, err
	}
	agrupadas := make(map[string]*RuaAgrupada)
	pendentes := 0
	for indice := range objetos {
		objeto := &objetos[indice]
		if objeto.StatusResolucao == models.StatusResolucaoPendente {
			if resolucao, erroResolucao := s.resolver.Resolver(ctx, objeto.TextoEntrada); erroResolucao == nil {
				if resolucao.Status == models.StatusResolucaoIdentificado {
					// Reprocessa pendências antigas quando uma melhoria do
					// resolvedor passa a identificá-las com segurança.
					objeto.RuaID = resolucao.RuaID
					objeto.NomeRua = resolucao.NomeRua
					objeto.ChaveAgrupamento = resolucao.ChaveAgrupamento
					objeto.Numero = resolucao.Numero
					objeto.CEP = resolucao.CEP
					objeto.StatusResolucao = resolucao.Status
					objeto.MotivoPendencia = resolucao.MotivoPendencia
					objeto.OpcoesResolucao = nil
					if s.coordenadas != nil {
						coordenada, erroCoordenada := s.coordenadas.Resolver(ctx, SolicitacaoCoordenada{
							ChaveAgrupamento: resolucao.ChaveAgrupamento,
							NomeRua:          resolucao.NomeRua,
							RuaID:            resolucao.RuaID,
							PermitirExterno:  true,
						})
						if erroCoordenada != nil {
							slog.Warn("não foi possível obter coordenada da pendência reprocessada", "error", erroCoordenada)
						} else if coordenada != nil {
							objeto.Latitude = &coordenada.Latitude
							objeto.Longitude = &coordenada.Longitude
							objeto.FonteCoordenada = coordenada.Fonte
						}
					}
					if erroAtualizacao := s.repo.UpdateObjeto(ctx, objeto); erroAtualizacao != nil {
						slog.Warn("não foi possível salvar a resolução reprocessada", "error", erroAtualizacao)
					}
				} else {
					objeto.OpcoesResolucao = resolucao.Opcoes
				}
			}
		}
		if objeto.StatusResolucao != models.StatusResolucaoIdentificado || objeto.ChaveAgrupamento == "" {
			pendentes++
			continue
		}
		rua := agrupadas[objeto.ChaveAgrupamento]
		if rua == nil {
			rua = &RuaAgrupada{
				Chave: objeto.ChaveAgrupamento, NomeRua: objeto.NomeRua,
				Latitude: objeto.Latitude, Longitude: objeto.Longitude, FonteCoordenada: objeto.FonteCoordenada,
			}
			agrupadas[objeto.ChaveAgrupamento] = rua
		}
		if rua.Latitude == nil && objeto.Latitude != nil && objeto.Longitude != nil {
			rua.Latitude = objeto.Latitude
			rua.Longitude = objeto.Longitude
			rua.FonteCoordenada = objeto.FonteCoordenada
		}
		rua.Quantidade++
	}
	ruas := make([]RuaAgrupada, 0, len(agrupadas))
	semCoordenadas := 0
	for _, rua := range agrupadas {
		if rua.Latitude == nil && s.coordenadas != nil {
			coordenada, erroCoordenada := s.coordenadas.Resolver(ctx, SolicitacaoCoordenada{
				ChaveAgrupamento: rua.Chave,
				NomeRua:          rua.NomeRua,
				PermitirExterno:  false,
			})
			if erroCoordenada != nil {
				slog.Warn("não foi possível consultar coordenada interna da rua", "error", erroCoordenada)
			} else if coordenada != nil {
				rua.Latitude = &coordenada.Latitude
				rua.Longitude = &coordenada.Longitude
				rua.FonteCoordenada = coordenada.Fonte
			}
		}
		if rua.Latitude == nil || rua.Longitude == nil {
			semCoordenadas++
		}
		ruas = append(ruas, *rua)
	}
	sort.Slice(ruas, func(i, j int) bool { return ruas[i].NomeRua < ruas[j].NomeRua })

	detalhe := &OrdenamentoDetalhe{
		ID:                  ordenamento.ID,
		Status:              ordenamento.Status,
		CreatedAt:           ordenamento.CreatedAt,
		UpdatedAt:           ordenamento.UpdatedAt,
		TotalObjetos:        len(objetos),
		TotalRuas:           len(ruas),
		TotalPendentes:      pendentes,
		TotalSemCoordenadas: semCoordenadas,
		Objetos:             objetos,
		Ruas:                ruas,
	}
	if s.paradas != nil {
		ordemSugerida, err := s.paradas.ListByOrdenamento(ctx, ordenamento.ID)
		if err != nil {
			return nil, err
		}
		detalhe.OrdemSugerida = ordemSugerida
	}
	return detalhe, nil
}
