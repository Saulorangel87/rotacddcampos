package services

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/empresa/rotas-entrega/models"
)

func novaCargaMemoria(t *testing.T) (OrdenamentoService, *ordenamentoRepoFake, *paradaOrdenamentoRepoFake) {
	t.Helper()
	lat, lon := -21.75, -41.32
	paradas := &paradaOrdenamentoRepoFake{}
	repo := &ordenamentoRepoFake{
		ativo: &models.Ordenamento{ID: 7, UsuarioID: 42, Status: models.StatusOrdenamentoEmAndamento}, paradas: paradas,
		objetos: []models.ObjetoOrdenamento{
			{ID: 1, OrdenamentoID: 7, NomeRua: "RUA A", ChaveAgrupamento: "rua:10", StatusResolucao: models.StatusResolucaoIdentificado, Latitude: &lat, Longitude: &lon},
			{ID: 2, OrdenamentoID: 7, NomeRua: "RUA B", ChaveAgrupamento: "rua:20", StatusResolucao: models.StatusResolucaoIdentificado, Latitude: &lat, Longitude: &lon},
		},
	}
	s := NewOrdenamentoService(repo, enderecoResolverFake{}, nil, paradas, otimizadorFake{})
	if _, err := s.GerarOrdem(context.Background(), 42, 7, GerarOrdemDTO{}); err != nil {
		t.Fatal(err)
	}
	return s, repo, paradas
}

func salvarInvertida(t *testing.T, s OrdenamentoService, p *paradaOrdenamentoRepoFake, reutilizar bool) *OrdenamentoDetalhe {
	t.Helper()
	ids := []uint{p.paradas[1].ID, p.paradas[0].ID}
	d, err := s.SalvarOrdemFinal(context.Background(), 42, 7, SalvarOrdemFinalDTO{ParadaIDs: ids, Reutilizar: reutilizar})
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func sequenciaFinal(d *OrdenamentoDetalhe) []string {
	paradas := append([]models.ParadaOrdenamento(nil), d.OrdemSugerida...)
	sort.Slice(paradas, func(i, j int) bool {
		a, b := paradas[i].OrdemSugerida, paradas[j].OrdemSugerida
		if paradas[i].OrdemFinal != nil {
			a = *paradas[i].OrdemFinal
		}
		if paradas[j].OrdemFinal != nil {
			b = *paradas[j].OrdemFinal
		}
		return a < b
	})
	var chaves []string
	for _, p := range paradas {
		chaves = append(chaves, p.ChaveAgrupamento)
	}
	return chaves
}

func TestMemoriaPreservaAjusteAoGerarNovamente(t *testing.T) {
	s, _, p := novaCargaMemoria(t)
	salvo := salvarInvertida(t, s, p, false)
	antigoID := p.paradas[0].ID
	gerado, err := s.GerarOrdem(context.Background(), 42, 7, GerarOrdemDTO{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(sequenciaFinal(gerado), sequenciaFinal(salvo)) {
		t.Fatal("gerar novamente perdeu o ajuste")
	}
	if gerado.OrdemSugerida[0].ID == antigoID {
		t.Fatal("teste deve exercitar novos IDs de paradas")
	}
	if len(p.historico) != 1 {
		t.Fatal("geração automática não deve criar correções humanas")
	}
	if p.historico[0].Paradas[0].OrdemSugerida != 2 {
		t.Fatal("histórico perdeu a sugestão original")
	}
	consultado, err := s.GetAtivo(context.Background(), 42)
	if err != nil || !reflect.DeepEqual(sequenciaFinal(consultado), sequenciaFinal(salvo)) {
		t.Fatal("reconsulta perdeu ordem final", err)
	}
}

func TestMemoriaPessoalSobreviveLimpezaSemEnsinarAjustePontual(t *testing.T) {
	for _, reutilizar := range []bool{false, true} {
		t.Run(map[bool]string{false: "carga", true: "pessoal"}[reutilizar], func(t *testing.T) {
			s, repo, p := novaCargaMemoria(t)
			objetos := append([]models.ObjetoOrdenamento(nil), repo.objetos...)
			salvo := salvarInvertida(t, s, p, reutilizar)
			vazio, err := s.Limpar(context.Background(), 42, 7)
			if err != nil || len(vazio.OrdemSugerida) != 0 || len(vazio.HistoricoOrdens) != 1 {
				t.Fatal("limpeza deve manter histórico e remover carga", err)
			}
			// Outra carga, com quantidade e IDs operacionais diferentes.
			repo.objetos = append(objetos, objetos[0])
			repo.objetos[2].ID = 3
			gerado, err := s.GerarOrdem(context.Background(), 42, 7, GerarOrdemDTO{})
			if err != nil {
				t.Fatal(err)
			}
			igual := reflect.DeepEqual(sequenciaFinal(gerado), sequenciaFinal(salvo))
			if igual != reutilizar {
				t.Fatalf("reutilização inesperada: %v", sequenciaFinal(gerado))
			}
			if gerado.TotalObjetos != 3 {
				t.Fatal("memória não pode restaurar objetos da carga antiga")
			}
			if reutilizar && gerado.OrdemSugerida[0].FonteOrdemFinal != "pessoal" {
				t.Fatal("falta origem da sequência reutilizada")
			}
		})
	}
}

func TestMemoriaNaoAplicaAOutraPessoaAreaOuConjunto(t *testing.T) {
	for _, caso := range []string{"outro_usuario", "mesmo_nome_outro_cadastro", "rua_nova", "lista_parcial", "outra_origem"} {
		t.Run(caso, func(t *testing.T) {
			s, repo, p := novaCargaMemoria(t)
			salvarInvertida(t, s, p, true)
			p.paradas = nil
			usuarioID := uint(42)
			switch caso {
			case "outro_usuario":
				usuarioID = 99
				repo.ativo.UsuarioID = 99
			case "mesmo_nome_outro_cadastro":
				repo.objetos[0].ChaveAgrupamento = "rua:999"
			case "rua_nova":
				rua := repo.objetos[0]
				rua.ChaveAgrupamento = "rua:30"
				repo.objetos = append(repo.objetos, rua)
			case "lista_parcial":
				repo.objetos = repo.objetos[:1]
			case "outra_origem":
				p.historico[0].Contexto = "outro-cdd"
			}
			d, err := s.GerarOrdem(context.Background(), usuarioID, 7, GerarOrdemDTO{})
			if err != nil {
				t.Fatal(err)
			}
			for _, parada := range d.OrdemSugerida {
				if parada.OrdemFinal != nil {
					t.Fatal("aplicou memória incompatível")
				}
			}
			if caso == "outro_usuario" && len(d.HistoricoOrdens) != 0 {
				t.Fatal("expôs histórico de outro usuário")
			}
		})
	}
}

func TestMemoriaRecalculoExplicitoMantemHistoricoEReferencia(t *testing.T) {
	s, _, p := novaCargaMemoria(t)
	salvo := salvarInvertida(t, s, p, true)
	d, err := s.GerarOrdem(context.Background(), 42, 7, GerarOrdemDTO{SomenteAlgoritmo: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, parada := range d.OrdemSugerida {
		if parada.OrdemFinal != nil {
			t.Fatal("recalcular sem ajustes aplicou memória")
		}
	}
	if len(d.HistoricoOrdens) != 1 || d.ReferenciaPessoalID == nil {
		t.Fatal("recalcular apagou memória")
	}
	d, err = s.GerarOrdem(context.Background(), 42, 7, GerarOrdemDTO{})
	if err != nil || !reflect.DeepEqual(sequenciaFinal(d), sequenciaFinal(salvo)) {
		t.Fatal("geração normal não reaplicou memória", err)
	}
}

func TestMemoriaDesativaSemApagarHistoricoOuReativarVersaoAntiga(t *testing.T) {
	s, _, p := novaCargaMemoria(t)
	salvarInvertida(t, s, p, true)
	salvo := salvarInvertida(t, s, p, true)
	if p.historico[0].DesativadaEm == nil || len(p.historico) != 2 {
		t.Fatal("substituição deve manter versão anterior desativada")
	}
	id := *salvo.ReferenciaPessoalID
	if err := s.EsquecerReferencia(context.Background(), 99, id); !errors.Is(err, ErrReferenciaNaoEncontrada) {
		t.Fatal("outro usuário conseguiu desativar", err)
	}
	if err := s.EsquecerReferencia(context.Background(), 42, id); err != nil {
		t.Fatal(err)
	}
	p.paradas = nil
	d, err := s.GerarOrdem(context.Background(), 42, 7, GerarOrdemDTO{})
	if err != nil || len(d.HistoricoOrdens) != 2 || d.ReferenciaPessoalID != nil {
		t.Fatal("desativação inconsistente", err)
	}
	for _, parada := range d.OrdemSugerida {
		if parada.OrdemFinal != nil {
			t.Fatal("reativou versão antiga")
		}
	}
}

func TestMemoriaRejeitaDadosInvalidosSemGravarHistorico(t *testing.T) {
	for _, caso := range []string{"duplicada", "ausente", "estranha", "zero", "outro_dono", "trecho_sem_cadastro"} {
		t.Run(caso, func(t *testing.T) {
			s, repo, p := novaCargaMemoria(t)
			dto := SalvarOrdemFinalDTO{ParadaIDs: []uint{p.paradas[0].ID, p.paradas[1].ID}, Reutilizar: true}
			esperado := ErrOrdemFinalInvalida
			switch caso {
			case "duplicada":
				dto.ParadaIDs[1] = dto.ParadaIDs[0]
			case "ausente":
				dto.ParadaIDs = dto.ParadaIDs[:1]
			case "estranha":
				dto.ParadaIDs[1] = 999
			case "zero":
				dto.ParadaIDs[1] = 0
			case "outro_dono":
				repo.ativo.UsuarioID = 99
				esperado = ErrOrdenamentoNaoEncontrado
			case "trecho_sem_cadastro":
				p.paradas[0].ChaveAgrupamento = "nome:RUA A"
				esperado = ErrReferenciaSemCadastro
			}
			_, err := s.SalvarOrdemFinal(context.Background(), 42, 7, dto)
			if !errors.Is(err, esperado) || len(p.historico) != 0 {
				t.Fatal("aceitou dados inválidos", err)
			}
		})
	}
}
