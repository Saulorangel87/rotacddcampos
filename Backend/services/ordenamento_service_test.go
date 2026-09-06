package services

import (
	"context"
	"errors"
	"testing"

	"github.com/empresa/rotas-entrega/models"
)

type ordenamentoRepoFake struct {
	ativo     *models.Ordenamento
	objetos   []models.ObjetoOrdenamento
	erroBusca error
	erroCriar error
}

func (r *ordenamentoRepoFake) FindAtivoByUsuario(_ context.Context, _ uint) (*models.Ordenamento, error) {
	return r.ativo, r.erroBusca
}

func (r *ordenamentoRepoFake) Create(_ context.Context, ordenamento *models.Ordenamento) error {
	if r.erroCriar != nil {
		return r.erroCriar
	}
	ordenamento.ID = 1
	r.ativo = ordenamento
	return nil
}

func (r *ordenamentoRepoFake) ListObjetos(_ context.Context, _ uint) ([]models.ObjetoOrdenamento, error) {
	return r.objetos, nil
}

func (r *ordenamentoRepoFake) CreateObjeto(_ context.Context, objeto *models.ObjetoOrdenamento) error {
	objeto.ID = uint(len(r.objetos) + 1)
	r.objetos = append(r.objetos, *objeto)
	return nil
}

func (r *ordenamentoRepoFake) DeleteObjeto(_ context.Context, ordenamentoID, objetoID uint) (bool, error) {
	for i, objeto := range r.objetos {
		if objeto.ID == objetoID && objeto.OrdenamentoID == ordenamentoID {
			r.objetos = append(r.objetos[:i], r.objetos[i+1:]...)
			return true, nil
		}
	}
	return false, nil
}

type enderecoResolverFake struct {
	resolucao ResolucaoEndereco
}

type coordenadaResolverFake struct {
	coordenada *CoordenadaReferencia
}

type paradaOrdenamentoRepoFake struct {
	paradas []models.ParadaOrdenamento
}

func (r *paradaOrdenamentoRepoFake) ListByOrdenamento(_ context.Context, _ uint) ([]models.ParadaOrdenamento, error) {
	return append([]models.ParadaOrdenamento(nil), r.paradas...), nil
}

func (r *paradaOrdenamentoRepoFake) ReplaceByOrdenamento(_ context.Context, _ uint, paradas []models.ParadaOrdenamento) error {
	r.paradas = append([]models.ParadaOrdenamento(nil), paradas...)
	return nil
}

func (r *paradaOrdenamentoRepoFake) DeleteByOrdenamento(_ context.Context, _ uint) error {
	r.paradas = nil
	return nil
}

type otimizadorFake struct{}

func (otimizadorFake) Otimizar(_ PontoGeografico, paradas []ParadaParaOtimizar) []ParadaParaOtimizar {
	resultado := append([]ParadaParaOtimizar(nil), paradas...)
	for inicio, fim := 0, len(resultado)-1; inicio < fim; inicio, fim = inicio+1, fim-1 {
		resultado[inicio], resultado[fim] = resultado[fim], resultado[inicio]
	}
	return resultado
}

func (r coordenadaResolverFake) Resolver(_ context.Context, _ SolicitacaoCoordenada) (*CoordenadaReferencia, error) {
	return r.coordenada, nil
}

func (r enderecoResolverFake) Resolver(_ context.Context, _ string) (ResolucaoEndereco, error) {
	return r.resolucao, nil
}

func TestOrdenamentoServiceCriar(t *testing.T) {
	repo := &ordenamentoRepoFake{}
	service := NewOrdenamentoService(repo, enderecoResolverFake{}, nil, nil, nil)

	ordenamento, err := service.Criar(context.Background(), 42)
	if err != nil {
		t.Fatalf("Criar() erro inesperado: %v", err)
	}
	if repo.ativo.UsuarioID != 42 {
		t.Fatalf("UsuarioID = %d, esperado 42", repo.ativo.UsuarioID)
	}
	if ordenamento.Status != models.StatusOrdenamentoEmAndamento {
		t.Fatalf("Status = %q, esperado %q", ordenamento.Status, models.StatusOrdenamentoEmAndamento)
	}
}

func TestOrdenamentoServiceNaoCriaSegundoAtivo(t *testing.T) {
	repo := &ordenamentoRepoFake{ativo: &models.Ordenamento{ID: 7, UsuarioID: 42}}
	service := NewOrdenamentoService(repo, enderecoResolverFake{}, nil, nil, nil)

	_, err := service.Criar(context.Background(), 42)
	if !errors.Is(err, ErrOrdenamentoJaExiste) {
		t.Fatalf("Criar() erro = %v, esperado %v", err, ErrOrdenamentoJaExiste)
	}
}

func TestOrdenamentoServiceAgrupaObjetosPorRua(t *testing.T) {
	ruaID := uint(10)
	repo := &ordenamentoRepoFake{ativo: &models.Ordenamento{ID: 7, UsuarioID: 42, Status: models.StatusOrdenamentoEmAndamento}}
	resolver := enderecoResolverFake{resolucao: ResolucaoEndereco{
		RuaID:            &ruaID,
		NomeRua:          "AVENIDA SETE DE SETEMBRO",
		ChaveAgrupamento: "nome:SETE DE SETEMBRO",
		Status:           models.StatusResolucaoIdentificado,
	}}
	service := NewOrdenamentoService(repo, resolver, nil, nil, nil)

	for i := 0; i < 2; i++ {
		if _, err := service.AdicionarObjeto(context.Background(), 42, 7, AdicionarObjetoDTO{Entrada: "Sete de Setembro", Origem: "manual"}); err != nil {
			t.Fatalf("AdicionarObjeto() erro inesperado: %v", err)
		}
	}
	detalhe, err := service.GetAtivo(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetAtivo() erro inesperado: %v", err)
	}
	if detalhe.TotalObjetos != 2 || detalhe.TotalRuas != 1 || detalhe.Ruas[0].Quantidade != 2 {
		t.Fatalf("resumo inesperado: %+v", detalhe)
	}
}

func TestOrdenamentoServiceNaoAcessaOrdenamentoDeOutroUsuario(t *testing.T) {
	repo := &ordenamentoRepoFake{ativo: &models.Ordenamento{ID: 9, UsuarioID: 99, Status: models.StatusOrdenamentoEmAndamento}}
	service := NewOrdenamentoService(repo, enderecoResolverFake{}, nil, nil, nil)

	_, err := service.AdicionarObjeto(context.Background(), 42, 7, AdicionarObjetoDTO{Entrada: "Rua A"})
	if !errors.Is(err, ErrOrdenamentoNaoEncontrado) {
		t.Fatalf("AdicionarObjeto() erro = %v, esperado %v", err, ErrOrdenamentoNaoEncontrado)
	}
}

func TestOrdenamentoServiceGuardaCoordenadaDoObjeto(t *testing.T) {
	ruaID := uint(10)
	repo := &ordenamentoRepoFake{ativo: &models.Ordenamento{ID: 7, UsuarioID: 42, Status: models.StatusOrdenamentoEmAndamento}}
	resolverEndereco := enderecoResolverFake{resolucao: ResolucaoEndereco{
		RuaID: &ruaID, NomeRua: "RUA TESTE", ChaveAgrupamento: "rua:10", Status: models.StatusResolucaoIdentificado,
	}}
	resolverCoordenada := coordenadaResolverFake{coordenada: &CoordenadaReferencia{
		Latitude: -21.75, Longitude: -41.32, Fonte: models.FonteCoordenadaGeometria,
	}}
	service := NewOrdenamentoService(repo, resolverEndereco, resolverCoordenada, nil, nil)

	detalhe, err := service.AdicionarObjeto(context.Background(), 42, 7, AdicionarObjetoDTO{Entrada: "Rua Teste"})
	if err != nil {
		t.Fatalf("AdicionarObjeto() erro inesperado: %v", err)
	}
	objeto := detalhe.Objetos[0]
	if objeto.Latitude == nil || objeto.Longitude == nil || *objeto.Latitude != -21.75 || *objeto.Longitude != -41.32 {
		t.Fatalf("coordenada do objeto inesperada: %+v", objeto)
	}
	if detalhe.TotalSemCoordenadas != 0 || detalhe.Ruas[0].FonteCoordenada != models.FonteCoordenadaGeometria {
		t.Fatalf("resumo de coordenadas inesperado: %+v", detalhe)
	}
}

func TestOrdenamentoServiceGeraOrdemSugerida(t *testing.T) {
	latitudeA, longitudeA := -21.75, -41.32
	latitudeB, longitudeB := -21.76, -41.31
	repo := &ordenamentoRepoFake{
		ativo: &models.Ordenamento{ID: 7, UsuarioID: 42, Status: models.StatusOrdenamentoEmAndamento},
		objetos: []models.ObjetoOrdenamento{
			{ID: 1, OrdenamentoID: 7, NomeRua: "RUA A", ChaveAgrupamento: "rua:a", StatusResolucao: models.StatusResolucaoIdentificado, Latitude: &latitudeA, Longitude: &longitudeA},
			{ID: 2, OrdenamentoID: 7, NomeRua: "RUA B", ChaveAgrupamento: "rua:b", StatusResolucao: models.StatusResolucaoIdentificado, Latitude: &latitudeB, Longitude: &longitudeB},
		},
	}
	paradas := &paradaOrdenamentoRepoFake{}
	service := NewOrdenamentoService(repo, enderecoResolverFake{}, nil, paradas, otimizadorFake{})

	detalhe, err := service.GerarOrdem(context.Background(), 42, 7)
	if err != nil {
		t.Fatalf("GerarOrdem() erro inesperado: %v", err)
	}
	if len(paradas.paradas) != 2 || len(detalhe.OrdemSugerida) != 2 {
		t.Fatalf("ordem sugerida inesperada: %+v", detalhe.OrdemSugerida)
	}
	if detalhe.OrdemSugerida[0].NomeRua != "RUA B" || detalhe.OrdemSugerida[0].OrdemSugerida != 1 {
		t.Fatalf("primeira parada inesperada: %+v", detalhe.OrdemSugerida[0])
	}
}

func TestOrdenamentoServiceNaoGeraOrdemComPendencia(t *testing.T) {
	repo := &ordenamentoRepoFake{
		ativo:   &models.Ordenamento{ID: 7, UsuarioID: 42, Status: models.StatusOrdenamentoEmAndamento},
		objetos: []models.ObjetoOrdenamento{{ID: 1, OrdenamentoID: 7, TextoEntrada: "Rua inexistente", StatusResolucao: models.StatusResolucaoPendente}},
	}
	service := NewOrdenamentoService(repo, enderecoResolverFake{}, nil, &paradaOrdenamentoRepoFake{}, otimizadorFake{})

	_, err := service.GerarOrdem(context.Background(), 42, 7)
	if !errors.Is(err, ErrOrdenamentoComPendencias) {
		t.Fatalf("GerarOrdem() erro = %v, esperado %v", err, ErrOrdenamentoComPendencias)
	}
}

func TestOrdenamentoServiceNaoGeraOrdemSemCoordenada(t *testing.T) {
	repo := &ordenamentoRepoFake{
		ativo:   &models.Ordenamento{ID: 7, UsuarioID: 42, Status: models.StatusOrdenamentoEmAndamento},
		objetos: []models.ObjetoOrdenamento{{ID: 1, OrdenamentoID: 7, NomeRua: "RUA A", ChaveAgrupamento: "rua:a", StatusResolucao: models.StatusResolucaoIdentificado}},
	}
	service := NewOrdenamentoService(repo, enderecoResolverFake{}, nil, &paradaOrdenamentoRepoFake{}, otimizadorFake{})

	_, err := service.GerarOrdem(context.Background(), 42, 7)
	if !errors.Is(err, ErrOrdenamentoSemCoordenadas) {
		t.Fatalf("GerarOrdem() erro = %v, esperado %v", err, ErrOrdenamentoSemCoordenadas)
	}
}
