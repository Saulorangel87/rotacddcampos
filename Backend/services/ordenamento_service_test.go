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

func (r enderecoResolverFake) Resolver(_ context.Context, _ string) (ResolucaoEndereco, error) {
	return r.resolucao, nil
}

func TestOrdenamentoServiceCriar(t *testing.T) {
	repo := &ordenamentoRepoFake{}
	service := NewOrdenamentoService(repo, enderecoResolverFake{})

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
	service := NewOrdenamentoService(repo, enderecoResolverFake{})

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
	service := NewOrdenamentoService(repo, resolver)

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
	service := NewOrdenamentoService(repo, enderecoResolverFake{})

	_, err := service.AdicionarObjeto(context.Background(), 42, 7, AdicionarObjetoDTO{Entrada: "Rua A"})
	if !errors.Is(err, ErrOrdenamentoNaoEncontrado) {
		t.Fatalf("AdicionarObjeto() erro = %v, esperado %v", err, ErrOrdenamentoNaoEncontrado)
	}
}
