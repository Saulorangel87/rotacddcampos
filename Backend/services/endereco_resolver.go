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
		"TRAVESSA": true, "TV": true, "PRACA": true, "ESTRADA": true,
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
}

type EnderecoResolver interface {
	Resolver(ctx context.Context, entrada string) (ResolucaoEndereco, error)
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
	if len(grupos) == 0 {
		return ResolucaoEndereco{}, false, nil
	}
	if len(grupos) > 1 {
		return pendencia(MotivoRuaAmbigua, numero), true, nil
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
		return pendencia(MotivoRuaAmbigua, numero), true, nil
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

func pendencia(motivo, numero string) ResolucaoEndereco {
	return ResolucaoEndereco{
		Numero:          numero,
		Status:          models.StatusResolucaoPendente,
		MotivoPendencia: motivo,
	}
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
	return strings.TrimSpace(reSufixoSegmento.ReplaceAllString(nome, ""))
}

func normalizarNomeBase(texto string) string {
	texto = nomeBaseExibicao(texto)
	texto = normalizarTexto(texto)
	partes := strings.Fields(texto)
	if len(partes) > 0 && prefixosLogradouro[partes[0]] {
		partes = partes[1:]
	}
	if len(partes) > 0 && prefixosLogradouro[partes[len(partes)-1]] {
		partes = partes[:len(partes)-1]
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
