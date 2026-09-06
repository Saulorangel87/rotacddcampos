package services

import (
	"context"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/empresa/rotas-entrega/models"
	"golang.org/x/text/unicode/norm"
)

const (
	MotivoRuaNaoEncontrada = "rua_nao_encontrada"
	MotivoRuaAmbigua       = "rua_ambigua"
)

var (
	reEspacos          = regexp.MustCompile(`\s+`)
	reNaoAlfanumerico  = regexp.MustCompile(`[^A-Z0-9]+`)
	reNumeroFinal      = regexp.MustCompile(`(?i)(?:,\s*|\s+(?:N(?:[º°O.]|UMERO)?\s*)?)(\d+[A-Z]?)\s*$`)
	reSufixoSegmento   = regexp.MustCompile(`(?i)\s+-\s+(AT[ÉE]|DE|LADO)\s+.*$`)
	reFaixaAte         = regexp.MustCompile(`\bATE\s+(\d+)`)
	reFaixaDe          = regexp.MustCompile(`\bDE\s+(\d+)\s+(?:A|AO)\s+(\d+|FIM)\b`)
	prefixosLogradouro = map[string]bool{
		"RUA": true, "R": true, "AVENIDA": true, "AV": true,
		"TRAVESSA": true, "TV": true, "PRACA": true, "PCA": true, "ESTRADA": true,
		"RODOVIA": true, "ALAMEDA": true, "LARGO": true, "BOULEVARD": true,
	}
	palavrasBuscaIgnoradas = map[string]bool{
		"A": true, "AS": true, "O": true, "OS": true, "DA": true, "DAS": true,
		"DE": true, "DO": true, "DOS": true, "E": true,
	}
)

type ruaBuscaRepository interface {
	FindAll(ctx context.Context, filters map[string]string) ([]models.Rua, error)
}

type ResolucaoEndereco struct {
	RuaID            *uint
	NomeRua          string
	ChaveAgrupamento string
	Numero           string
	CEP              string
	Status           string
	MotivoPendencia  string
	Opcoes           []models.OpcaoResolucao
}

type EnderecoResolver interface {
	Resolver(ctx context.Context, entrada string) (ResolucaoEndereco, error)
}

// EnderecoResolverSelecionavel é implementado pelo resolvedor real para
// confirmar uma opção escolhida pelo usuário sem ampliar a interface usada
// pelos demais consumidores e pelos testes unitários.
type EnderecoResolverSelecionavel interface {
	Selecionar(ctx context.Context, entrada string, ruaID uint) (ResolucaoEndereco, error)
}

type enderecoResolver struct {
	ruas ruaBuscaRepository
}

func NewEnderecoResolver(ruas ruaBuscaRepository) EnderecoResolver {
	return &enderecoResolver{ruas: ruas}
}

func (r *enderecoResolver) Resolver(ctx context.Context, entrada string) (ResolucaoEndereco, error) {
	entrada = strings.TrimSpace(entrada)
	if entrada == "" {
		return pendencia(MotivoRuaNaoEncontrada, ""), nil
	}

	// Primeiro preserva números que realmente fazem parte do nome, como "Rua 13".
	if resolucao, encontrou, err := r.resolverTermo(ctx, entrada, ""); err != nil || encontrou {
		return resolucao, err
	}

	nome, numero := extrairNumeroFinal(entrada)
	if numero == "" {
		return pendencia(MotivoRuaNaoEncontrada, ""), nil
	}
	resolucao, encontrou, err := r.resolverTermo(ctx, nome, numero)
	if err != nil {
		return ResolucaoEndereco{}, err
	}
	if !encontrou {
		return pendencia(MotivoRuaNaoEncontrada, numero), nil
	}
	return resolucao, nil
}

func (r *enderecoResolver) resolverTermo(ctx context.Context, termo, numero string) (ResolucaoEndereco, bool, error) {
	normalizado := normalizarNomeBase(termo)
	if normalizado == "" {
		return ResolucaoEndereco{}, false, nil
	}

	candidatas, err := r.ruas.FindAll(ctx, map[string]string{"nome": melhorTokenBusca(normalizado)})
	if err != nil {
		return ResolucaoEndereco{}, false, err
	}

	gruposExatos := make(map[string][]models.Rua)
	gruposParciais := make(map[string][]models.Rua)
	for _, rua := range candidatas {
		base := normalizarNomeBase(rua.NomeRua)
		if base == normalizado {
			gruposExatos[base] = append(gruposExatos[base], rua)
			continue
		}
		if len(normalizado) >= 4 && strings.Contains(base, normalizado) {
			gruposParciais[base] = append(gruposParciais[base], rua)
		}
	}
	// Uma rua cujo nome normalizado coincide por completo é mais confiável do
	// que um nome maior que apenas contém o termo. Por exemplo, "SILVA
	// TAVARES" não deve ficar ambígua por também existir "PEDREVAL DA SILVA
	// TAVARES" no cadastro.
	grupos := gruposParciais
	if len(gruposExatos) > 0 {
		grupos = gruposExatos
	}
	// Quando a pessoa informa o tipo do logradouro, use essa informação para
	// separar ruas homônimas. Ex.: no cadastro há uma "TRAVESSA SÃO GONÇALO"
	// e três "RUA SÃO GONÇALO". O tipo não faz parte da chave de agrupamento,
	// mas é uma evidência segura para escolher o grupo correto.
	if tipo := tipoLogradouroInformado(termo); tipo != "" {
		gruposPorTipo := make(map[string][]models.Rua)
		for base, ruas := range grupos {
			for _, rua := range ruas {
				if tipoLogradouroNormalizado(rua.NomeRua) == tipo {
					gruposPorTipo[base] = append(gruposPorTipo[base], rua)
				}
			}
		}
		// Só aplicar o filtro quando ele encontrar candidatos. Assim, uma
		// abreviação ausente no cadastro não transforma uma busca válida em
		// "não encontrada".
		if len(gruposPorTipo) > 0 {
			grupos = gruposPorTipo
		}
	}
	if len(grupos) == 0 {
		return ResolucaoEndereco{}, false, nil
	}
	if len(grupos) > 1 {
		return pendenciaComOpcoes(MotivoRuaAmbigua, numero, grupos), true, nil
	}

	var base string
	var ruas []models.Rua
	for base, ruas = range grupos {
	}

	segmentada := false
	for _, rua := range ruas {
		if temSufixoSegmento(rua.NomeRua) {
			segmentada = true
			break
		}
	}
	if len(ruas) > 1 && !segmentada {
		return pendenciaComOpcoes(MotivoRuaAmbigua, numero, grupos), true, nil
	}

	selecionada := selecionarTrecho(ruas, numero)
	nomeRua := nomeBaseExibicao(ruas[0].NomeRua)
	resolucao := ResolucaoEndereco{
		NomeRua:          nomeRua,
		ChaveAgrupamento: "nome:" + base,
		Numero:           numero,
		Status:           models.StatusResolucaoIdentificado,
	}
	if selecionada != nil {
		id := selecionada.ID
		resolucao.RuaID = &id
		resolucao.CEP = selecionada.CEP
		if !segmentada {
			resolucao.ChaveAgrupamento = "rua:" + strconv.FormatUint(uint64(id), 10)
		}
	}
	return resolucao, true, nil
}

// Selecionar confirma um cadastro específico dentre as opções de uma busca
// ambígua. A validação repete a normalização e nunca aceita uma rua que não
// corresponda ao texto informado.
func (r *enderecoResolver) Selecionar(ctx context.Context, entrada string, ruaID uint) (ResolucaoEndereco, error) {
	entrada = strings.TrimSpace(entrada)
	if entrada == "" || ruaID == 0 {
		return pendencia(MotivoRuaNaoEncontrada, ""), nil
	}

	for _, tentativa := range []struct {
		nome   string
		numero string
	}{
		{nome: entrada},
		func() struct {
			nome   string
			numero string
		} {
			nome, numero := extrairNumeroFinal(entrada)
			return struct {
				nome   string
				numero string
			}{nome: nome, numero: numero}
		}(),
	} {
		if tentativa.nome == "" {
			continue
		}
		normalizado := normalizarNomeBase(tentativa.nome)
		if normalizado == "" {
			continue
		}
		candidatas, err := r.ruas.FindAll(ctx, map[string]string{"nome": melhorTokenBusca(normalizado)})
		if err != nil {
			return ResolucaoEndereco{}, err
		}
		tipo := tipoLogradouroInformado(tentativa.nome)
		for _, rua := range candidatas {
			baseCandidata := normalizarNomeBase(rua.NomeRua)
			if rua.ID != ruaID || !nomeCandidatoCorresponde(baseCandidata, normalizado) {
				continue
			}
			if tipo != "" && tipoLogradouroNormalizado(rua.NomeRua) != tipo {
				continue
			}
			id := rua.ID
			return ResolucaoEndereco{
				RuaID:            &id,
				NomeRua:          nomeBaseExibicao(rua.NomeRua),
				ChaveAgrupamento: "rua:" + strconv.FormatUint(uint64(id), 10),
				Numero:           tentativa.numero,
				CEP:              rua.CEP,
				Status:           models.StatusResolucaoIdentificado,
			}, nil
		}
		if tentativa.numero != "" {
			break
		}
	}
	return pendencia(MotivoRuaNaoEncontrada, ""), nil
}

func nomeCandidatoCorresponde(candidato, termo string) bool {
	if candidato == termo {
		return true
	}
	return len(termo) >= 4 && strings.Contains(candidato, termo)
}

func pendencia(motivo, numero string) ResolucaoEndereco {
	return ResolucaoEndereco{
		Numero:          numero,
		Status:          models.StatusResolucaoPendente,
		MotivoPendencia: motivo,
	}
}

func pendenciaComOpcoes(motivo, numero string, grupos map[string][]models.Rua) ResolucaoEndereco {
	resultado := pendencia(motivo, numero)
	vistos := make(map[uint]bool)
	for _, ruas := range grupos {
		for _, rua := range ruas {
			if vistos[rua.ID] {
				continue
			}
			vistos[rua.ID] = true
			resultado.Opcoes = append(resultado.Opcoes, models.OpcaoResolucao{
				RuaID: rua.ID, NomeRua: nomeBaseExibicao(rua.NomeRua), Distrito: rua.Distrito, CEP: rua.CEP,
			})
		}
	}
	sort.SliceStable(resultado.Opcoes, func(i, j int) bool {
		if resultado.Opcoes[i].NomeRua != resultado.Opcoes[j].NomeRua {
			return resultado.Opcoes[i].NomeRua < resultado.Opcoes[j].NomeRua
		}
		if resultado.Opcoes[i].Distrito != resultado.Opcoes[j].Distrito {
			return resultado.Opcoes[i].Distrito < resultado.Opcoes[j].Distrito
		}
		return resultado.Opcoes[i].RuaID < resultado.Opcoes[j].RuaID
	})
	return resultado
}

func extrairNumeroFinal(entrada string) (string, string) {
	maior := reNumeroFinal.FindStringSubmatchIndex(entrada)
	if maior == nil {
		return strings.TrimSpace(entrada), ""
	}
	numero := strings.ToUpper(entrada[maior[2]:maior[3]])
	nome := strings.TrimSpace(entrada[:maior[0]])
	return nome, numero
}

func selecionarTrecho(ruas []models.Rua, numeroTexto string) *models.Rua {
	if len(ruas) == 1 {
		return &ruas[0]
	}
	if numeroTexto == "" {
		return nil
	}
	numero, err := strconv.Atoi(strings.TrimRight(numeroTexto, "ABCDEFGHIJKLMNOPQRSTUVWXYZ"))
	if err != nil {
		return nil
	}

	var compativeis []models.Rua
	for _, rua := range ruas {
		if trechoAceitaNumero(rua.NomeRua, numero) {
			compativeis = append(compativeis, rua)
		}
	}
	if len(compativeis) == 1 {
		return &compativeis[0]
	}
	return nil
}

func trechoAceitaNumero(nome string, numero int) bool {
	n := normalizarTexto(nome)
	if strings.Contains(n, "LADO PAR") && numero%2 != 0 {
		return false
	}
	if strings.Contains(n, "LADO IMPAR") && numero%2 == 0 {
		return false
	}
	if partes := reFaixaDe.FindStringSubmatch(n); len(partes) > 0 {
		inicio, _ := strconv.Atoi(partes[1])
		if numero < inicio {
			return false
		}
		if partes[2] != "FIM" {
			fim, _ := strconv.Atoi(partes[2])
			return numero <= fim
		}
		return true
	}
	if partes := reFaixaAte.FindStringSubmatch(n); len(partes) > 0 {
		fim, _ := strconv.Atoi(partes[1])
		return numero <= fim
	}
	return true
}

func temSufixoSegmento(nome string) bool {
	return reSufixoSegmento.MatchString(strings.ReplaceAll(nome, "–", "-"))
}

func nomeBaseExibicao(nome string) string {
	nome = strings.ReplaceAll(nome, "–", "-")
	nome = strings.TrimSpace(reSufixoSegmento.ReplaceAllString(nome, ""))

	// Alguns registros importados usam "NOME, RUA" ou "NOME, TV.SÃO".
	// Mantemos a grafia original das palavras, mas exibimos a forma usual do
	// logradouro para a encomenda e para a busca externa.
	partesVirgula := strings.SplitN(nome, ",", 2)
	if len(partesVirgula) != 2 {
		return nome
	}
	parteNome := strings.TrimSpace(partesVirgula[0])
	parteTipo := strings.TrimSpace(strings.ReplaceAll(partesVirgula[1], ".", " "))
	palavrasTipo := strings.Fields(parteTipo)
	if len(palavrasTipo) == 0 {
		return nome
	}
	tipo := tipoLogradouroExibicao(palavrasTipo[0])
	if tipo == "" {
		return nome
	}
	resto := strings.TrimSpace(strings.Join(palavrasTipo[1:], " "))
	if resto == "" {
		return tipo + " " + parteNome
	}
	return strings.TrimSpace(tipo + " " + resto + " " + parteNome)
}

func tipoLogradouroExibicao(token string) string {
	switch normalizarTexto(token) {
	case "R", "RUA":
		return "RUA"
	case "AV", "AVENIDA":
		return "AVENIDA"
	case "TV", "TRAVESSA":
		return "TRAVESSA"
	case "PCA", "PRACA":
		return "PRAÇA"
	case "ESTRADA":
		return "ESTRADA"
	case "RODOVIA":
		return "RODOVIA"
	case "ALAMEDA":
		return "ALAMEDA"
	case "LARGO":
		return "LARGO"
	case "BOULEVARD":
		return "BOULEVARD"
	default:
		return ""
	}
}

func tipoLogradouroNormalizado(nome string) string {
	palavras := strings.Fields(normalizarTexto(nomeBaseExibicao(nome)))
	if len(palavras) == 0 {
		return ""
	}
	if tipo := tipoLogradouroExibicao(palavras[0]); tipo != "" {
		return normalizarTexto(tipo)
	}
	return ""
}

func tipoLogradouroInformado(termo string) string {
	palavras := strings.Fields(normalizarTexto(termo))
	for _, palavra := range palavras {
		if tipo := tipoLogradouroExibicao(palavra); tipo != "" {
			return normalizarTexto(tipo)
		}
	}
	return ""
}

func normalizarNomeBase(texto string) string {
	texto = nomeBaseExibicao(texto)
	texto = normalizarTexto(texto)
	partes := strings.Fields(texto)
	if len(partes) == 0 {
		return ""
	}

	// O cadastro contém tanto "RUA NOME" quanto formatos importados como
	// "NOME, RUA". Há ainda registros em que o tipo fica entre as partes do
	// nome ("CECÍLIA, RUA SANTA"). Remover o tipo e, nesse último caso,
	// recolocar o trecho posterior antes do nome evita que a mesma rua tenha
	// duas chaves de agrupamento.
	indiceTipo := -1
	for indice, parte := range partes {
		if prefixosLogradouro[parte] {
			indiceTipo = indice
			break
		}
	}
	if indiceTipo == 0 {
		partes = partes[1:]
	} else if indiceTipo == len(partes)-1 {
		partes = partes[:len(partes)-1]
	} else if indiceTipo > 0 {
		reordenadas := append([]string{}, partes[indiceTipo+1:]...)
		reordenadas = append(reordenadas, partes[:indiceTipo]...)
		partes = reordenadas
	}
	return strings.Join(partes, " ")
}

func normalizarTexto(texto string) string {
	texto = strings.ToUpper(strings.ReplaceAll(texto, "–", "-"))
	decomposto := norm.NFD.String(texto)
	texto = strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) {
			return -1
		}
		return r
	}, decomposto)
	texto = reNaoAlfanumerico.ReplaceAllString(texto, " ")
	return strings.TrimSpace(reEspacos.ReplaceAllString(texto, " "))
}

func melhorTokenBusca(normalizado string) string {
	partes := strings.Fields(normalizado)
	filtradas := partes[:0]
	for _, parte := range partes {
		if !palavrasBuscaIgnoradas[parte] {
			filtradas = append(filtradas, parte)
		}
	}
	if len(filtradas) == 0 {
		return normalizado
	}
	sort.SliceStable(filtradas, func(i, j int) bool { return len(filtradas[i]) > len(filtradas[j]) })
	return filtradas[0]
}
