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

func TestEnderecoResolverPriorizaCorrespondenciaExataSobreParcial(t *testing.T) {
	resolver := NewEnderecoResolver(ruaBuscaRepoFake{ruas: []models.Rua{
		{ID: 306, NomeRua: "RUA PEDREVAL DA SILVA TAVARES", Distrito: "605"},
		{ID: 468, NomeRua: "RUA SILVA TAVARES", Distrito: "607", CEP: "28027080"},
	}})

	resultado, err := resolver.Resolver(context.Background(), "Rua Silva Tavares")
	if err != nil {
		t.Fatalf("Resolver() erro inesperado: %v", err)
	}
	if resultado.Status != models.StatusResolucaoIdentificado || resultado.RuaID == nil || *resultado.RuaID != 468 {
		t.Fatalf("resolução inesperada: %+v", resultado)
	}
	if resultado.NomeRua != "RUA SILVA TAVARES" || resultado.CEP != "28027080" {
		t.Fatalf("rua escolhida inesperada: %+v", resultado)
	}
}

func TestEnderecoResolverAceitaArtigosOmitidosEAbreviacao(t *testing.T) {
	resolver := NewEnderecoResolver(ruaBuscaRepoFake{ruas: []models.Rua{
		{ID: 75, NomeRua: "AVENIDA SETE DE SETEMBRO", Distrito: "614", CEP: "28010562"},
	}})

	resultado, err := resolver.Resolver(context.Background(), "Av. Sete Setembro")
	if err != nil {
		t.Fatalf("Resolver() erro inesperado: %v", err)
	}
	if resultado.Status != models.StatusResolucaoIdentificado || resultado.RuaID == nil || *resultado.RuaID != 75 {
		t.Fatalf("rua não identificada com artigo omitido: %+v", resultado)
	}
}

func TestEnderecoResolverExigeTodosOsTermosDaBusca(t *testing.T) {
	resolver := NewEnderecoResolver(ruaBuscaRepoFake{ruas: []models.Rua{
		{ID: 1, NomeRua: "RUA SILVA TAVARES", Distrito: "607"},
		{ID: 2, NomeRua: "RUA SILVA TORRES", Distrito: "608"},
	}})

	resultado, err := resolver.Resolver(context.Background(), "Silva Tavares")
	if err != nil {
		t.Fatalf("Resolver() erro inesperado: %v", err)
	}
	if resultado.Status != models.StatusResolucaoIdentificado || resultado.RuaID == nil || *resultado.RuaID != 1 {
		t.Fatalf("a busca não priorizou os dois termos informados: %+v", resultado)
	}
}

func TestEnderecoResolverAceitaPrefixoSemEscolherOutroTermo(t *testing.T) {
	resolver := NewEnderecoResolver(ruaBuscaRepoFake{ruas: []models.Rua{
		{ID: 1, NomeRua: "RUA SÉRGIO CARDOSO", Distrito: "608"},
		{ID: 2, NomeRua: "RUA SILVIO FONTOURA", Distrito: "608"},
	}})

	resultado, err := resolver.Resolver(context.Background(), "Serg")
	if err != nil {
		t.Fatalf("Resolver() erro inesperado: %v", err)
	}
	if resultado.Status != models.StatusResolucaoIdentificado || resultado.RuaID == nil || *resultado.RuaID != 1 {
		t.Fatalf("prefixo foi resolvido incorretamente: %+v", resultado)
	}
}

func TestEnderecoResolverAceitaTiposDeLogradouroInvertidos(t *testing.T) {
	tests := []struct {
		entrada  string
		cadastro models.Rua
		esperado string
	}{
		{entrada: "Rua Santa Cecilia", cadastro: models.Rua{ID: 1, NomeRua: "CECÍLIA, RUA SANTA"}, esperado: "RUA SANTA CECÍLIA"},
		{entrada: "Travessa Sao Goncalo", cadastro: models.Rua{ID: 2, NomeRua: "GONÇALO, TV.SÃO"}, esperado: "TRAVESSA SÃO GONÇALO"},
		{entrada: "Praca Athaide Barbosa", cadastro: models.Rua{ID: 3, NomeRua: "ATHAÍDE BARBOSA, PÇA"}, esperado: "PRAÇA ATHAÍDE BARBOSA"},
		{entrada: "Estrada Alto da Pitanga", cadastro: models.Rua{ID: 4, NomeRua: "ALTO DA PITANGA, EST."}, esperado: "ESTRADA ALTO DA PITANGA"},
	}

	for _, caso := range tests {
		resolver := NewEnderecoResolver(ruaBuscaRepoFake{ruas: []models.Rua{caso.cadastro}})
		resultado, err := resolver.Resolver(context.Background(), caso.entrada+" 10")
		if err != nil {
			t.Fatalf("Resolver(%q) erro inesperado: %v", caso.entrada, err)
		}
		if resultado.Status != models.StatusResolucaoIdentificado || resultado.RuaID == nil || *resultado.RuaID != caso.cadastro.ID {
			t.Fatalf("Resolver(%q) não identificou o cadastro: %+v", caso.entrada, resultado)
		}
		if resultado.NomeRua != caso.esperado {
			t.Fatalf("Resolver(%q) nome = %q, esperado %q", caso.entrada, resultado.NomeRua, caso.esperado)
		}
	}
}

func TestEnderecoResolverUsaTipoParaDesempatarRuasHomônimas(t *testing.T) {
	resolver := NewEnderecoResolver(ruaBuscaRepoFake{ruas: []models.Rua{
		{ID: 1109, NomeRua: "GONÇALO, TV.SÃO", Distrito: "614"},
		{ID: 197, NomeRua: "RUA SÃO GONÇALO", Distrito: "604"},
		{ID: 1809, NomeRua: "RUA SÃO GONÇALO", Distrito: "620"},
		{ID: 1924, NomeRua: "RUA SÃO GONÇALO", Distrito: "621"},
	}})

	resultado, err := resolver.Resolver(context.Background(), "Travessa Sao Goncalo")
	if err != nil {
		t.Fatalf("Resolver() erro inesperado: %v", err)
	}
	if resultado.Status != models.StatusResolucaoIdentificado || resultado.RuaID == nil || *resultado.RuaID != 1109 {
		t.Fatalf("resolução inesperada: %+v", resultado)
	}
}

func TestEnderecoResolverExpõeOpcoesParaRuaAmbigua(t *testing.T) {
	resolver := NewEnderecoResolver(ruaBuscaRepoFake{ruas: []models.Rua{
		{ID: 197, NomeRua: "RUA SÃO GONÇALO", Distrito: "604", CEP: "28023592"},
		{ID: 1809, NomeRua: "RUA SÃO GONÇALO", Distrito: "620", CEP: "28070216"},
	}})

	resultado, err := resolver.Resolver(context.Background(), "São Gonçalo")
	if err != nil {
		t.Fatalf("Resolver() erro inesperado: %v", err)
	}
	if resultado.Status != models.StatusResolucaoPendente || len(resultado.Opcoes) != 2 {
		t.Fatalf("opções inesperadas: %+v", resultado)
	}

	selecionavel, ok := resolver.(EnderecoResolverSelecionavel)
	if !ok {
		t.Fatal("resolvedor não permite selecionar opção")
	}
	confirmada, err := selecionavel.Selecionar(context.Background(), "São Gonçalo", 1809)
	if err != nil {
		t.Fatalf("Selecionar() erro inesperado: %v", err)
	}
	if confirmada.Status != models.StatusResolucaoIdentificado || confirmada.RuaID == nil || *confirmada.RuaID != 1809 || confirmada.CEP != "28070216" {
		t.Fatalf("seleção inesperada: %+v", confirmada)
	}
}

func TestEnderecoResolverConfirmaOpcaoDeCorrespondenciaParcial(t *testing.T) {
	resolver := NewEnderecoResolver(ruaBuscaRepoFake{ruas: []models.Rua{
		{ID: 1, NomeRua: "RODOVIA SÉRGIO BARROSO", Distrito: "604"},
		{ID: 2, NomeRua: "RUA SÉRGIO BUARQUE DE HOLANDA", Distrito: "608"},
		{ID: 3, NomeRua: "RUA SÉRGIO CARDOSO", Distrito: "608"},
	}})

	resultado, err := resolver.Resolver(context.Background(), "sergio")
	if err != nil {
		t.Fatalf("Resolver() erro inesperado: %v", err)
	}
	if resultado.Status != models.StatusResolucaoPendente || len(resultado.Opcoes) != 3 {
		t.Fatalf("opções parciais inesperadas: %+v", resultado)
	}

	selecionavel := resolver.(EnderecoResolverSelecionavel)
	confirmada, err := selecionavel.Selecionar(context.Background(), "sergio", 3)
	if err != nil {
		t.Fatalf("Selecionar() erro inesperado: %v", err)
	}
	if confirmada.Status != models.StatusResolucaoIdentificado || confirmada.RuaID == nil || *confirmada.RuaID != 3 {
		t.Fatalf("seleção parcial inesperada: %+v", confirmada)
	}
}

func TestEnderecoResolverConfirmaOpcaoExibidaMesmoComTipoGenerico(t *testing.T) {
	resolver := NewEnderecoResolver(ruaBuscaRepoFake{ruas: []models.Rua{
		{ID: 131, NomeRua: "AVENIDA MIGUEL RINALDI", Distrito: "614", CEP: "28148042"},
		{ID: 132, NomeRua: "TRAVESSA MIGUEL RINALDI", Distrito: "614", CEP: "28148014"},
	}})

	resultado, err := resolver.Resolver(context.Background(), "Rua Miguel Rinaldi")
	if err != nil {
		t.Fatalf("Resolver() erro inesperado: %v", err)
	}
	if resultado.Status != models.StatusResolucaoPendente || len(resultado.Opcoes) != 2 {
		t.Fatalf("opções inesperadas: %+v", resultado)
	}

	selecionavel := resolver.(EnderecoResolverSelecionavel)
	confirmada, err := selecionavel.Selecionar(context.Background(), "Rua Miguel Rinaldi", 131)
	if err != nil {
		t.Fatalf("Selecionar() erro inesperado: %v", err)
	}
	if confirmada.Status != models.StatusResolucaoIdentificado || confirmada.RuaID == nil || *confirmada.RuaID != 131 {
		t.Fatalf("seleção inesperada: %+v", confirmada)
	}
	if confirmada.NomeRua != "AVENIDA MIGUEL RINALDI" || confirmada.CEP != "28148042" {
		t.Fatalf("cadastro selecionado inesperado: %+v", confirmada)
	}

	naoExibida, err := selecionavel.Selecionar(context.Background(), "Rua Miguel Rinaldi", 999)
	if err != nil {
		t.Fatalf("Selecionar() para opção não exibida retornou erro inesperado: %v", err)
	}
	if naoExibida.Status == models.StatusResolucaoIdentificado {
		t.Fatalf("cadastro não exibido foi aceito: %+v", naoExibida)
	}
}

func TestEnderecoResolverIdentificaRuaPorCEP(t *testing.T) {
	resolver := NewEnderecoResolver(ruaBuscaRepoFake{ruas: []models.Rua{
		{ID: 75, NomeRua: "AVENIDA SETE DE SETEMBRO - DE 230 AO 490 - LADO PAR", Distrito: "614", CEP: "28010562"},
		{ID: 76, NomeRua: "RUA OUTRA", Distrito: "614", CEP: "28010561"},
	}})

	resultado, err := resolver.Resolver(context.Background(), "28010-562")
	if err != nil {
		t.Fatalf("Resolver() erro inesperado: %v", err)
	}
	if resultado.Status != models.StatusResolucaoIdentificado || resultado.RuaID == nil || *resultado.RuaID != 75 {
		t.Fatalf("resolução por CEP inesperada: %+v", resultado)
	}
	if resultado.NomeRua != "AVENIDA SETE DE SETEMBRO" || resultado.CEP != "28010562" {
		t.Fatalf("rua identificada por CEP inesperada: %+v", resultado)
	}
}

func TestEnderecoResolverDeixaCEPNaoCadastradoPendente(t *testing.T) {
	resolver := NewEnderecoResolver(ruaBuscaRepoFake{ruas: []models.Rua{
		{ID: 1, NomeRua: "RUA TESTE", CEP: "28010000"},
	}})

	resultado, err := resolver.Resolver(context.Background(), "99999-999")
	if err != nil {
		t.Fatalf("Resolver() erro inesperado: %v", err)
	}
	if resultado.Status != models.StatusResolucaoPendente || resultado.MotivoPendencia != MotivoCEPNaoEncontrado {
		t.Fatalf("pendência de CEP inesperada: %+v", resultado)
	}
}

func TestEnderecoResolverExpõeOpcoesParaCEPAmbiguo(t *testing.T) {
	resolver := NewEnderecoResolver(ruaBuscaRepoFake{ruas: []models.Rua{
		{ID: 10, NomeRua: "RUA SÃO GONÇALO", Distrito: "604", CEP: "28023592"},
		{ID: 11, NomeRua: "RUA SÃO GONÇALO", Distrito: "620", CEP: "28023592"},
	}})

	resultado, err := resolver.Resolver(context.Background(), "28023592")
	if err != nil {
		t.Fatalf("Resolver() erro inesperado: %v", err)
	}
	if resultado.Status != models.StatusResolucaoPendente || resultado.MotivoPendencia != MotivoRuaAmbigua || len(resultado.Opcoes) != 2 {
		t.Fatalf("opções de CEP ambíguo inesperadas: %+v", resultado)
	}

	selecionavel := resolver.(EnderecoResolverSelecionavel)
	confirmada, err := selecionavel.Selecionar(context.Background(), "28023592", 11)
	if err != nil {
		t.Fatalf("Selecionar() erro inesperado: %v", err)
	}
	if confirmada.Status != models.StatusResolucaoIdentificado || confirmada.RuaID == nil || *confirmada.RuaID != 11 {
		t.Fatalf("seleção por CEP inesperada: %+v", confirmada)
	}
}
