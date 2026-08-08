// Comando standalone que importa o saldo em aberto de folgas a partir da
// planilha "CONTROLE DE FOLGAS - CDD CPS e GRU" (uma aba por colaborador).
//
// A planilha tem anos de histórico já consumido (saldo restante 0). Por
// pedido explícito, só entra aqui o que ainda está em aberto: uma linha
// abaixo por lançamento com saldo restante > 0 na planilha, cada uma virando
// um crédito com o motivo e a data originais (quando a data dava pra ler com
// segurança — texto solto tipo "CDD GUARUS" ficou sem data, só motivo e
// quantidade mesmo). Não é um número resumido por colaborador: é o extrato
// real, exatamente como ficaria se cada um desses créditos tivesse sido
// lançado pelo sistema na época.
//
// Rodar uma vez só (idempotente: se um lançamento com a mesma
// matrícula+motivo+data já existir, ele é pulado — rodar de novo não duplica):
//
//	go run cmd/importar-folgas-iniciais/main.go
//
// Use -confirmar pra gravar de verdade; sem essa flag só mostra o que
// seria feito (dry-run), pra conferir a lista antes de gravar.
//
//	go run cmd/importar-folgas-iniciais/main.go -confirmar
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/empresa/rotas-entrega/config"
	"github.com/empresa/rotas-entrega/database"
	"github.com/empresa/rotas-entrega/models"
)

type lancamentoInicial struct {
	Matricula  string
	Nome       string
	Motivo     string
	Data       string // DD/MM/AAAA, pode vir vazio quando a planilha não tinha data legível
	Quantidade int
}

// Extraído da planilha em 08/08/2026 — só linhas com saldo restante > 0.
var lancamentosIniciais = []lancamentoInicial{
	{Matricula: "89609565", Nome: "ALDIONE DE CARVALHO RIBEIRO", Motivo: "FERIADO", Data: "06/08/2026", Quantidade: 2},
	{Matricula: "89560744", Nome: "ALESSANDRO PEREIRA LOUZAN", Motivo: "TRE", Data: "05/10/2024", Quantidade: 2},
	{Matricula: "89560744", Nome: "ALESSANDRO PEREIRA LOUZAN", Motivo: "TRE", Data: "06/10/2024", Quantidade: 2},
	{Matricula: "89611047", Nome: "ALEXANDRE LOPES DE MORAES", Motivo: "CDD GUARUS", Data: "", Quantidade: 3},
	{Matricula: "89564332", Nome: "ANDREI FACINI GOULART COLODETTE", Motivo: "CDD GUARUS", Data: "", Quantidade: 5},
	{Matricula: "89606477", Nome: "BRUNO DE SOUZA SANTOS", Motivo: "FERIADO", Data: "06/08/2026", Quantidade: 3},
	{Matricula: "89606477", Nome: "BRUNO DE SOUZA SANTOS", Motivo: "Folga concedida pelo apoio ao CEE Campos", Data: "20/12/2025", Quantidade: 1},
	{Matricula: "89612779", Nome: "BRUNO PINTO FONSECA", Motivo: "CDD GUARUS", Data: "", Quantidade: 3},
	{Matricula: "83241205", Nome: "CARLOS JONAS SOARES LOPES", Motivo: "CDD GUARUS", Data: "", Quantidade: 2},
	{Matricula: "83241205", Nome: "CARLOS JONAS SOARES LOPES", Motivo: "FERIADO", Data: "06/08/2026", Quantidade: 3},
	{Matricula: "83241205", Nome: "CARLOS JONAS SOARES LOPES", Motivo: "Niver", Data: "", Quantidade: 1},
	{Matricula: "89518225", Nome: "CLÁUDIO LUIZ LYRIO BARRETO", Motivo: "CDD GUARUS", Data: "", Quantidade: 2},
	{Matricula: "89611730", Nome: "DOUGLAS DOS SANTOS CARVALHO", Motivo: "FERIADO", Data: "06/08/2026", Quantidade: 2},
	{Matricula: "89611730", Nome: "DOUGLAS DOS SANTOS CARVALHO", Motivo: "PPP", Data: "26/07/2023", Quantidade: 1},
	{Matricula: "89609689", Nome: "ERLÂNDIA GARCIA DA SILVA", Motivo: "FERIADO", Data: "06/08/2026", Quantidade: 3},
	{Matricula: "89609689", Nome: "ERLÂNDIA GARCIA DA SILVA", Motivo: "TRE 2024", Data: "06/10/2024", Quantidade: 2},
	{Matricula: "89609689", Nome: "ERLÂNDIA GARCIA DA SILVA", Motivo: "aniversario", Data: "10/07/2026", Quantidade: 1},
	{Matricula: "89611691", Nome: "FABIO ROBERTO RAMOS REIS JUNIOR", Motivo: "FERIADO", Data: "06/08/2026", Quantidade: 2},
	{Matricula: "89611691", Nome: "FABIO ROBERTO RAMOS REIS JUNIOR", Motivo: "Niver", Data: "12/06/2026", Quantidade: 1},
	{Matricula: "89611691", Nome: "FABIO ROBERTO RAMOS REIS JUNIOR", Motivo: "RT", Data: "23/04/2024", Quantidade: 2},
	{Matricula: "89611691", Nome: "FABIO ROBERTO RAMOS REIS JUNIOR", Motivo: "RT", Data: "01/09/2024", Quantidade: 2},
	{Matricula: "89611691", Nome: "FABIO ROBERTO RAMOS REIS JUNIOR", Motivo: "TFSP", Data: "10/08/2024", Quantidade: 1},
	{Matricula: "89613325", Nome: "FABIO VIANA SOUZA", Motivo: "CDD GUARUS", Data: "", Quantidade: 6},
	{Matricula: "86026879", Nome: "FENELON LOPES DA SILVA", Motivo: "Abono aniversário", Data: "24/08/2025", Quantidade: 1},
	{Matricula: "89622260", Nome: "GUSTAVO JOSE MOREIRA ALVARENGA", Motivo: "CDD GUARUS", Data: "", Quantidade: 5},
	{Matricula: "89609573", Nome: "IGOR DA SILVA BARBOSA", Motivo: "FERIADO", Data: "06/08/2026", Quantidade: 2},
	{Matricula: "89611829", Nome: "JANSEM GONÇALVES GOMES", Motivo: "CDD GUARUS", Data: "", Quantidade: 1},
	{Matricula: "89556640", Nome: "JOAO BATISTA ALVES DE SOUZA JUNIOR", Motivo: "CDD GUARUS", Data: "13/07/2026", Quantidade: 2},
	{Matricula: "89592255", Nome: "JOÃO VICENTE DOS SANTOS MACHADO", Motivo: "FERIADO", Data: "06/08/2026", Quantidade: 2},
	{Matricula: "89581393", Nome: "JULIANA ANDRADE NOVAES FRANÇA", Motivo: "FERIADO", Data: "06/08/2026", Quantidade: 2},
	{Matricula: "89620941", Nome: "LANDERSON LOPES GONÇALVES", Motivo: "CDD GUARUS", Data: "", Quantidade: 1},
	{Matricula: "89585119", Nome: "LAURA GOMES DOS SANTOS PESSANHA SANTIAGO", Motivo: "Apoio CDD Campos", Data: "", Quantidade: 2},
	{Matricula: "89606469", Nome: "LUCAS BRAGA TEODORO", Motivo: "TRE 2024", Data: "06/10/2024", Quantidade: 1},
	{Matricula: "89557115", Nome: "LUIZ CARLOS DE FREITAS CARLOS", Motivo: "FERIADO", Data: "06/08/2026", Quantidade: 2},
	{Matricula: "89557115", Nome: "LUIZ CARLOS DE FREITAS CARLOS", Motivo: "RT", Data: "07/06/2026", Quantidade: 2},
	{Matricula: "89557115", Nome: "LUIZ CARLOS DE FREITAS CARLOS", Motivo: "RT", Data: "14/07/2026", Quantidade: 2},
	{Matricula: "89557115", Nome: "LUIZ CARLOS DE FREITAS CARLOS", Motivo: "TFSP", Data: "13/07/2026", Quantidade: 1},
	{Matricula: "89557115", Nome: "LUIZ CARLOS DE FREITAS CARLOS", Motivo: "TRE 2024", Data: "05/10/2024", Quantidade: 1},
	{Matricula: "89557115", Nome: "LUIZ CARLOS DE FREITAS CARLOS", Motivo: "TRE 2024", Data: "06/10/2024", Quantidade: 2},
	{Matricula: "89609794", Nome: "MAIKE BASTOS ROCHA", Motivo: "FERIADO", Data: "06/08/2026", Quantidade: 2},
	{Matricula: "83248099", Nome: "MARCO ANTONIO SALGADO DE SOUZA", Motivo: "CDD GUARUS", Data: "", Quantidade: 6},
	{Matricula: "83248099", Nome: "MARCO ANTONIO SALGADO DE SOUZA", Motivo: "FERIADO", Data: "06/08/2026", Quantidade: 2},
	{Matricula: "89609638", Nome: "MARCUS VINICIUS MANHÃES SEPULVIDA", Motivo: "CDD GUARUS", Data: "", Quantidade: 2},
	{Matricula: "83208666", Nome: "MAURO PINTO DE JESUS", Motivo: "FERIADO", Data: "06/08/2026", Quantidade: 2},
	{Matricula: "83208666", Nome: "MAURO PINTO DE JESUS", Motivo: "Folgas concedidas por ter realizado reparos na unidade", Data: "", Quantidade: 1},
	{Matricula: "89622090", Nome: "MAYCK NOGUEIRA AMORIM", Motivo: "TRE", Data: "06/10/2024", Quantidade: 2},
	{Matricula: "89523997", Nome: "MICHELLY VOLOTÃO SOUZA FERREIRA", Motivo: "FERIADO", Data: "", Quantidade: 2},
	{Matricula: "89523997", Nome: "MICHELLY VOLOTÃO SOUZA FERREIRA", Motivo: "Folga concedida pelo Apoio ao CEE Campos", Data: "21/12/2025", Quantidade: 1},
	{Matricula: "89525906", Nome: "PAULO CÉSAR SILVA DOS SANTOS", Motivo: "TRE 2024", Data: "06/10/2024", Quantidade: 1},
	{Matricula: "89609670", Nome: "PAULO RENATO DA SILVA GUIMARÃES", Motivo: "FERIADO", Data: "06/08/2026", Quantidade: 2},
	{Matricula: "89609670", Nome: "PAULO RENATO DA SILVA GUIMARÃES", Motivo: "RT", Data: "15/01/2024", Quantidade: 2},
	{Matricula: "89609670", Nome: "PAULO RENATO DA SILVA GUIMARÃES", Motivo: "RT", Data: "25/02/2024", Quantidade: 2},
	{Matricula: "89560833", Nome: "ROSEMIL VIDAL CAMPOS", Motivo: "Aniversario", Data: "03/07/2025", Quantidade: 1},
	{Matricula: "89610903", Nome: "SAULO RANGEL ROSA LEONARDO", Motivo: "Apoio ao gestor no SD", Data: "18/04/2026", Quantidade: 1},
	{Matricula: "89610903", Nome: "SAULO RANGEL ROSA LEONARDO", Motivo: "FERIADO", Data: "06/08/2026", Quantidade: 2},
	{Matricula: "89530810", Nome: "SÉRGIO CAMPOS RODRIGUES", Motivo: "Folga", Data: "30/09/2024", Quantidade: 1},
	{Matricula: "89530810", Nome: "SÉRGIO CAMPOS RODRIGUES", Motivo: "Folga", Data: "25/10/2024", Quantidade: 1},
	{Matricula: "89530810", Nome: "SÉRGIO CAMPOS RODRIGUES", Motivo: "PPP 2024", Data: "16/07/2024", Quantidade: 1},
	{Matricula: "89530810", Nome: "SÉRGIO CAMPOS RODRIGUES", Motivo: "RT", Data: "15/01/2024", Quantidade: 2},
	{Matricula: "89578392", Nome: "VINÍCIUS DE SOUZA RODRIGUES", Motivo: "FERIADO", Data: "06/08/2026", Quantidade: 2},
	{Matricula: "89578392", Nome: "VINÍCIUS DE SOUZA RODRIGUES", Motivo: "Folga concedida pelo apoio ao CEE Campos", Data: "", Quantidade: 2},
}

const sufixoImportacao = "(importado da planilha)"

func main() {
	confirmar := flag.Bool("confirmar", false, "grava de verdade; sem essa flag só mostra o que seria feito")
	flag.Parse()

	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		slog.Error("falha ao conectar no banco", "error", err)
		os.Exit(1)
	}

	if err := database.RunMigrations(db); err != nil {
		slog.Error("falha ao rodar migrations", "error", err)
		os.Exit(1)
	}

	var totalGravado, totalPulado int
	naoEncontrados := map[string]string{}

	for _, item := range lancamentosIniciais {
		var colaborador models.Colaborador
		if err := db.Where("matricula = ?", item.Matricula).First(&colaborador).Error; err != nil {
			naoEncontrados[item.Matricula] = item.Nome
			continue
		}

		// A data entra no motivo de propósito: várias pessoas têm o mesmo
		// motivo repetido em datas diferentes (ex: dois "RT"), e sem a data
		// no texto o check de duplicidade abaixo trataria o segundo como
		// "já importado" e pularia ele, perdendo um crédito de verdade.
		motivo := item.Motivo + " " + sufixoImportacao
		if item.Data != "" {
			motivo = item.Motivo + " (" + item.Data + ") " + sufixoImportacao
		}

		var jaExiste int64
		db.Model(&models.FolgaLancamento{}).
			Where("matricula = ? AND motivo = ?", item.Matricula, motivo).
			Count(&jaExiste)
		if jaExiste > 0 {
			fmt.Printf("PULADO  %s (%s): %s — já importado\n", item.Matricula, item.Nome, item.Motivo)
			totalPulado++
			continue
		}

		fmt.Printf("%s  %s (%s): +%d — %s%s\n", acao(*confirmar), item.Matricula, item.Nome, item.Quantidade, item.Motivo, dataOuVazio(item.Data))

		if !*confirmar {
			totalGravado++
			continue
		}

		lancamento := models.FolgaLancamento{
			Matricula:  item.Matricula,
			Tipo:       models.FolgaCredito,
			Quantidade: item.Quantidade,
			Motivo:     motivo,
			CriadoPor:  "importacao-planilha",
		}
		if item.Data != "" {
			data, err := time.Parse("02/01/2006", item.Data)
			if err == nil {
				lancamento.DataReferencia = &data
			}
		}
		if err := db.Create(&lancamento).Error; err != nil {
			slog.Error("falha ao gravar lançamento", "matricula", item.Matricula, "motivo", item.Motivo, "error", err)
			continue
		}

		db.Create(&models.HistoricoAlteracao{
			Tipo:      "folga",
			Descricao: fmt.Sprintf("Lançou +%d folga(s) para matrícula %s — %s", item.Quantidade, item.Matricula, motivo),
			Usuario:   "importacao-planilha",
		})

		totalGravado++
	}

	fmt.Println()
	fmt.Printf("Total: %d lançados, %d pulados (já importados).\n", totalGravado, totalPulado)
	if len(naoEncontrados) > 0 {
		fmt.Println()
		fmt.Println("Matrículas da planilha que NÃO foram encontradas na tabela de colaboradores (nada foi gravado pra elas):")
		for matricula, nome := range naoEncontrados {
			fmt.Printf(" - %s (%s)\n", matricula, nome)
		}
	}
	if !*confirmar {
		fmt.Println()
		fmt.Println("Isso foi um dry-run — nada foi gravado. Rode com -confirmar pra gravar de verdade.")
	}
}

func acao(confirmar bool) string {
	if confirmar {
		return "GRAVANDO"
	}
	return "SIMULADO"
}

func dataOuVazio(data string) string {
	if data == "" {
		return ""
	}
	return " (" + data + ")"
}
