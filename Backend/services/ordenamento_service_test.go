package services

import (
	"context"
	"errors"
	"testing"

	"github.com/empresa/rotas-entrega/models"
)

type ordenamentoRepoFake struct {
	ativo     *models.Ordenamento
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

func TestOrdenamentoServiceCriar(t *testing.T) {
	repo := &ordenamentoRepoFake{}
	service := NewOrdenamentoService(repo)

	ordenamento, err := service.Criar(context.Background(), 42)
	if err != nil {
		t.Fatalf("Criar() erro inesperado: %v", err)
	}
	if ordenamento.UsuarioID != 42 {
		t.Fatalf("UsuarioID = %d, esperado 42", ordenamento.UsuarioID)
	}
	if ordenamento.Status != models.StatusOrdenamentoEmAndamento {
		t.Fatalf("Status = %q, esperado %q", ordenamento.Status, models.StatusOrdenamentoEmAndamento)
	}
}

func TestOrdenamentoServiceNaoCriaSegundoAtivo(t *testing.T) {
	repo := &ordenamentoRepoFake{ativo: &models.Ordenamento{ID: 7, UsuarioID: 42}}
	service := NewOrdenamentoService(repo)

	_, err := service.Criar(context.Background(), 42)
	if !errors.Is(err, ErrOrdenamentoJaExiste) {
		t.Fatalf("Criar() erro = %v, esperado %v", err, ErrOrdenamentoJaExiste)
	}
}
