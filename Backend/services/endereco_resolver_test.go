package services

import (
	"context"
	"testing"

	"github.com/empresa/rotas-entrega/models"
)

type ruaBuscaRepoFake struct {
	ruas []models.Rua
}

func (r ruaBuscaRepoFake) FindAll(_ context.Context, _ map[string]string) ([]models.Rua, error) {
	return r.ruas, nil
}

func TestEnderecoResolverSelecionaTrechoPorNumeroEParidade(t *testing.T) {
	resolver := NewEnderecoResolver(ruaBuscaRepoFake{ruas: []models.Rua{
		{ID: 134, NomeRua: "AVENIDA SETE DE SETEMBRO - ATÉ 221 - LADO ÍMPAR", CEP: "28013331"},
		{ID: 135, NomeRua: "AVENIDA SETE DE SETEMBRO - ATÉ 206 - LADO PAR", CEP: "28013332"},
		{ID: 74, NomeRua: "AVENIDA SETE DE SETEMBRO - DE 235 AO 509 - LADO ÍMPAR", CEP: "28010561"},
		{ID: 75, NomeRua: "AVENIDA SETE DE SETEMBRO - DE 230 AO 490 - LADO PAR", CEP: "28010562"},
	}})

	resultado, err := resolver.Resolver(context.Background(), "Avenida Sete de Setembro 342")
	if err != nil {
		t.Fatalf("Resolver() erro inesperado: %v", err)
	}
	if resultado.RuaID == nil || *resultado.RuaID != 75 {
		t.Fatalf("RuaID = %v, esperado 75", resultado.RuaID)
	}
	if resultado.Numero != "342" || resultado.NomeRua != "AVENIDA SETE DE SETEMBRO" {
		t.Fatalf("resolução inesperada: %+v", resultado)
	}
}

func TestEnderecoResolverAceitaRuaSegmentadaSemNumero(t *testing.T) {
	resolver := NewEnderecoResolver(ruaBuscaRepoFake{ruas: []models.Rua{
		{ID: 140, NomeRua: "RUA TENENTE-CORONEL CARDOSO - ATÉ 427 - LADO ÍMPAR"},
		{ID: 141, NomeRua: "RUA TENENTE-CORONEL CARDOSO - ATÉ 424 - LADO PAR"},
	}})

	resultado, err := resolver.Resolver(context.Background(), "Rua Tenente Coronel Cardoso")
	if err != nil {
		t.Fatalf("Resolver() erro inesperado: %v", err)
	}
	if resultado.Status != models.StatusResolucaoIdentificado || resultado.RuaID != nil {
		t.Fatalf("resolução inesperada: %+v", resultado)
	}
	if resultado.ChaveAgrupamento != "nome:TENENTE CORONEL CARDOSO" {
		t.Fatalf("chave = %q", resultado.ChaveAgrupamento)
	}
}

func TestEnderecoResolverMantemNomeComNumero(t *testing.T) {
	resolver := NewEnderecoResolver(ruaBuscaRepoFake{ruas: []models.Rua{
		{ID: 20, NomeRua: "RUA 13"},
	}})

	resultado, err := resolver.Resolver(context.Background(), "Rua 13")
	if err != nil {
		t.Fatalf("Resolver() erro inesperado: %v", err)
	}
	if resultado.RuaID == nil || *resultado.RuaID != 20 || resultado.Numero != "" {
		t.Fatalf("resolução inesperada: %+v", resultado)
	}
}

func TestEnderecoResolverDeixaNomeGenericoDuplicadoPendente(t *testing.T) {
	resolver := NewEnderecoResolver(ruaBuscaRepoFake{ruas: []models.Rua{
		{ID: 1, NomeRua: "RUA A", Distrito: "601"},
		{ID: 2, NomeRua: "RUA A", Distrito: "610"},
	}})

	resultado, err := resolver.Resolver(context.Background(), "Rua A 20")
	if err != nil {
		t.Fatalf("Resolver() erro inesperado: %v", err)
	}
	if resultado.Status != models.StatusResolucaoPendente || resultado.MotivoPendencia != MotivoRuaAmbigua {
		t.Fatalf("resolução inesperada: %+v", resultado)
	}
}

func TestEnderecoResolverDeixaRuaDesconhecidaPendente(t *testing.T) {
	resolver := NewEnderecoResolver(ruaBuscaRepoFake{})

	resultado, err := resolver.Resolver(context.Background(), "Rua que não existe 50")
	if err != nil {
		t.Fatalf("Resolver() erro inesperado: %v", err)
	}
	if resultado.Status != models.StatusResolucaoPendente || resultado.MotivoPendencia != MotivoRuaNaoEncontrada {
		t.Fatalf("resolução inesperada: %+v", resultado)
	}
}
