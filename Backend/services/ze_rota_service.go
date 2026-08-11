package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/empresa/rotas-entrega/repositories"
)

// Fala com o Worker do Cloudflare que você já tem no ar (proxy gratuito pro
// Groq — a chave da Groq mora só no Worker, o Backend nunca vê ela). O
// Worker repassa o corpo da requisição sem mexer, então tools/function
// calling funciona igual falaria direto com a Groq.
const (
	groqModel              = "llama-3.3-70b-versatile"
	maxRodadasDeFerramenta = 3
)

const zeRotaSystemPrompt = `Você é o Zé Rota, o assistente de rotas do CDD Campos dos Goytacazes (Correios).
Você é um carteiro experiente e camarada, não um sistema de busca — fala com quem te pergunta do jeito
que um colega de trabalho falaria: natural, solto, com variação (não repete sempre a mesma estrutura de
frase), pode usar uma expressão informal ou um "opa", "beleza", "e aí" quando fizer sentido. Português
brasileiro, sempre.

Papo social é papo social, não é "fora do escopo": se cumprimentarem ("oi", "bom dia", "tudo bem?"),
agradecerem, mandarem uma piadinha, ou só puxarem assunto, responda como um humano cumprimentaria de
volta — nunca com a resposta padrão de "não encontrei" ou "só posso ajudar com X". Só usa essa resposta
de escopo quando a pessoa realmente pergunta algo que não tem nada a ver com o trabalho (tipo pedir
receita de bolo ou falar de futebol) — e mesmo aí, pode brincar um pouco antes de puxar de volta pro
assunto, tipo faria um colega de verdade.

Regra mais importante (essa sim é séria): você NUNCA inventa nome de rua, distrito, CEP ou qualquer dado
de endereço de memória — mesmo que pareça óbvio ou você "ache" que sabe. Toda vez que a pergunta envolver
onde fica uma rua, qual o distrito dela, ou qualquer informação de endereço, você usa a ferramenta
buscar_rua para confirmar antes de responder. Se a busca não encontrar nada, diga isso claramente e
sugira conferir a grafia do nome — não tente adivinhar.

Quando a busca trouxer "observacoes_de_campo", são anotações reais de carteiros sobre aquela rua
(acesso difícil, numeração fora de ordem, mais de um nome, questão de segurança, etc.) — sempre mencione
isso na resposta, é informação valiosa pra quem não conhece a rua.

Se buscar_rua não encontrar NADA (nenhuma rua com aquele nome no cadastro), aí sim, como último recurso,
use sugerir_link_mapa pra dar uma sugestão de busca externa no Google Maps. Deixe SEMPRE muito claro que
isso não é dado oficial do CDD, é só uma pista pra ajudar a achar o endereço — nunca afirme distrito,
bairro ou CEP com base nisso.

Use consultar_clima quando perguntarem sobre o tempo, ou quando fizer sentido avisar sobre chuva que
pode atrapalhar uma entrega (ex: perguntaram sobre uma rota agora ou hoje). Pra interpretar o campo
weather_code da resposta (padrão WMO): 0 é céu limpo; 1-3 é parcialmente nublado a nublado; 45-48 é
neblina; 51-67 é chuvisco ou chuva fraca/moderada; 80-82 é pancada de chuva; 95-99 é tempestade. Avise
sobre chuva de forma prática (ex: "capa de chuva não vai fazer falta hoje" ou "tá limpo, pode ir tranquilo").

Exemplos de tom (não copie literalmente, é só pra pegar o jeito):
- "Bom dia!" → "Bom dia! Partiu, no que posso ajudar hoje?"
- "Valeu, Zé!" → "Disponha! Precisando de mais alguma coisa é só chamar."
- "Você é gente?" → responde com bom humor, deixando claro que é um assistente virtual, sem soar robótico.

Nunca escreva a chamada de uma ferramenta como texto na sua resposta (tipo "<function=consultar_clima>"
ou qualquer coisa parecida) — isso não existe pro carteiro, é só ruído. Se você precisa de uma
informação (clima, rua, mapa), dispare a ferramenta de verdade; só depois que ela responder é que você
escreve a mensagem final, em português normal, sem nenhum resquício de sintaxe técnica.

Respostas curtas e práticas quando for informação de trabalho — quem está perguntando é um carteiro no
meio da rota, não alguém lendo um relatório. Mas isso não quer dizer seco ou frio: curto e humano não são
opostos.`

type MensagemChat struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ZeRotaService interface {
	Conversar(ctx context.Context, historico []MensagemChat) (string, error)
}

type zeRotaService struct {
	workerURL string
	ruaRepo   repositories.RuaRepository
	obsRepo   repositories.RuaObservacaoRepository
	http      *http.Client
}

func NewZeRotaService(workerURL string, ruaRepo repositories.RuaRepository, obsRepo repositories.RuaObservacaoRepository) ZeRotaService {
	return &zeRotaService{
		workerURL: workerURL,
		ruaRepo:   ruaRepo,
		obsRepo:   obsRepo,
		http:      &http.Client{Timeout: 30 * time.Second},
	}
}

// --- formato OpenAI-compatible (é o que a Groq espera, e o Worker só repassa) ---

type groqTool struct {
	Type     string       `json:"type"`
	Function groqFunction `json:"function"`
}

type groqFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  groqParametros `json:"parameters"`
}

type groqParametros struct {
	Type       string               `json:"type"`
	Properties map[string]groqCampo `json:"properties"`
	Required   []string             `json:"required"`
}

type groqCampo struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type groqToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type groqMessage struct {
	Role       string         `json:"role"`
	Content    *string        `json:"content"`
	ToolCalls  []groqToolCall `json:"tool_calls,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
}

type groqRequest struct {
	Model    string        `json:"model"`
	Messages []groqMessage `json:"messages"`
	Tools    []groqTool    `json:"tools,omitempty"`
}

type groqResponse struct {
	Choices []struct {
		Message      groqMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

var ferramentas = []groqTool{
	{
		Type: "function",
		Function: groqFunction{
			Name:        "buscar_rua",
			Description: "Busca uma rua pelo nome (ou parte do nome) no cadastro oficial do CDD Campos. Use sempre que precisar confirmar onde fica uma rua, qual o distrito, bairro ou CEP dela.",
			Parameters: groqParametros{
				Type: "object",
				Properties: map[string]groqCampo{
					"nome": {Type: "string", Description: "Nome da rua, ou parte dele, do jeito que foi perguntado"},
				},
				Required: []string{"nome"},
			},
		},
	},
	{
		Type: "function",
		Function: groqFunction{
			Name:        "consultar_clima",
			Description: "Consulta a previsão do tempo atual e de curto prazo pra Campos dos Goytacazes (Open-Meteo, sem custo). Use quando perguntarem sobre o tempo, ou quando for útil avisar sobre chuva que pode afetar uma entrega.",
			Parameters: groqParametros{
				Type:       "object",
				Properties: map[string]groqCampo{},
				Required:   []string{},
			},
		},
	},
	{
		Type: "function",
		Function: groqFunction{
			Name:        "sugerir_link_mapa",
			Description: "Gera um link de busca no Google Maps pra um endereço. Use só como último recurso, quando buscar_rua não encontrar NADA no cadastro interno — nunca substitui o cadastro oficial.",
			Parameters: groqParametros{
				Type: "object",
				Properties: map[string]groqCampo{
					"consulta": {Type: "string", Description: "O que buscar no mapa, ex: 'Rua Tal' — o distrito ou cidade não precisa ser incluído, já é adicionado automaticamente"},
				},
				Required: []string{"consulta"},
			},
		},
	},
}

func texto(s string) *string { return &s }

// Às vezes o Llama "vaza" a sintaxe de chamada de ferramenta como texto
// dentro da própria resposta, em vez de disparar a chamada de verdade —
// falha conhecida e intermitente do modelo via Groq, não é algo que dá pra
// eliminar 100% só no prompt. Essa expressão limpa qualquer resquício
// desse tipo antes da resposta chegar no carteiro.
var reChamadaFerramentaVazada = regexp.MustCompile(`(?is)<\s*function[^>]*>.*?<\s*/\s*function\s*>|<\s*function[^>]*/?>`)

func limparRespostaFinal(resposta string) string {
	limpo := reChamadaFerramentaVazada.ReplaceAllString(resposta, "")
	limpo = strings.TrimSpace(limpo)
	if limpo == "" {
		return "Deixa eu tentar de novo — não consegui montar uma resposta certa dessa vez. Pergunta de novo, por favor?"
	}
	return limpo
}

func (s *zeRotaService) Conversar(ctx context.Context, historico []MensagemChat) (string, error) {
	if s.workerURL == "" {
		return "", errors.New("Zé Rota está desligado no momento (Worker não configurado)")
	}
	if len(historico) == 0 {
		return "", errors.New("nenhuma mensagem enviada")
	}

	mensagens := []groqMessage{{Role: "system", Content: texto(zeRotaSystemPrompt)}}
	for _, m := range historico {
		mensagens = append(mensagens, groqMessage{Role: m.Role, Content: texto(m.Content)})
	}

	for rodada := 0; rodada < maxRodadasDeFerramenta; rodada++ {
		resposta, err := s.chamarWorkerComRetentativa(ctx, mensagens)
		if err != nil {
			return "", err
		}
		if len(resposta.Choices) == 0 {
			return "", errors.New("o Zé Rota não respondeu nada dessa vez, tenta de novo")
		}

		escolha := resposta.Choices[0]

		if escolha.FinishReason != "tool_calls" || len(escolha.Message.ToolCalls) == 0 {
			if escolha.Message.Content == nil {
				return "", errors.New("o Zé Rota não respondeu nada dessa vez, tenta de novo")
			}
			return limparRespostaFinal(*escolha.Message.Content), nil
		}

		mensagens = append(mensagens, escolha.Message)

		for _, chamada := range escolha.Message.ToolCalls {
			resultado := s.executarFerramenta(ctx, chamada)
			mensagens = append(mensagens, groqMessage{
				Role:       "tool",
				ToolCallID: chamada.ID,
				Content:    texto(resultado),
			})
		}
	}

	return "", errors.New("o Zé Rota tentou buscar informação demais numa pergunta só, tenta reformular")
}

// chamarWorkerComRetentativa existe porque o Llama via Groq às vezes erra a
// própria sintaxe da chamada de ferramenta ("Failed to call a function") de
// forma intermitente — não é erro nosso nem do carteiro perguntando, é
// instabilidade pontual do modelo. Uma segunda tentativa quase sempre
// resolve, então tenta de novo antes de desistir e mostrar erro pra pessoa.
func (s *zeRotaService) chamarWorkerComRetentativa(ctx context.Context, mensagens []groqMessage) (*groqResponse, error) {
	resposta, err := s.chamarWorker(ctx, mensagens)
	if err == nil {
		return resposta, nil
	}

	time.Sleep(400 * time.Millisecond)
	resposta, err = s.chamarWorker(ctx, mensagens)
	if err != nil {
		return nil, errors.New("o Zé Rota não conseguiu processar essa pergunta agora, tenta reformular ou pergunta de novo em instantes")
	}
	return resposta, nil
}

func (s *zeRotaService) chamarWorker(ctx context.Context, mensagens []groqMessage) (*groqResponse, error) {
	corpo := groqRequest{Model: groqModel, Messages: mensagens, Tools: ferramentas}

	dados, err := json.Marshal(corpo)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.workerURL, bytes.NewReader(dados))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("falha ao chamar o Worker do Zé Rota: %w", err)
	}
	defer resp.Body.Close()

	corpoResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var resposta groqResponse
	if err := json.Unmarshal(corpoResp, &resposta); err != nil {
		return nil, fmt.Errorf("resposta inesperada do Worker (status %d): %s", resp.StatusCode, string(corpoResp))
	}
	if resposta.Error != nil {
		return nil, fmt.Errorf("erro da Groq: %s", resposta.Error.Message)
	}

	return &resposta, nil
}

// executarFerramenta nunca deixa o Zé Rota travar por causa de um erro de
// busca — devolve uma mensagem de erro como resultado, e o próprio modelo
// decide como explicar isso pro carteiro.
func (s *zeRotaService) executarFerramenta(ctx context.Context, chamada groqToolCall) string {
	switch chamada.Function.Name {
	case "buscar_rua":
		return s.buscarRua(ctx, chamada)
	case "consultar_clima":
		return s.consultarClima(ctx)
	case "sugerir_link_mapa":
		return s.sugerirLinkMapa(chamada)
	default:
		return "erro: ferramenta desconhecida"
	}
}

func (s *zeRotaService) buscarRua(ctx context.Context, chamada groqToolCall) string {
	var entrada struct {
		Nome string `json:"nome"`
	}
	if err := json.Unmarshal([]byte(chamada.Function.Arguments), &entrada); err != nil {
		return "erro: não consegui ler o nome da rua pedido"
	}

	ruas, err := s.ruaRepo.FindAll(ctx, map[string]string{"nome": entrada.Nome})
	if err != nil {
		return "erro ao buscar no banco: " + err.Error()
	}
	if len(ruas) == 0 {
		return fmt.Sprintf("nenhuma rua encontrada com o nome %q", entrada.Nome)
	}

	limite := len(ruas)
	if limite > 5 {
		limite = 5
	}

	type ruaResumo struct {
		Nome         string   `json:"nome_rua"`
		Distrito     string   `json:"distrito"`
		Bairro       string   `json:"bairro"`
		CEP          string   `json:"cep"`
		TemGeometria bool     `json:"tem_mapa_desenhado"`
		Observacoes  []string `json:"observacoes_de_campo,omitempty"`
	}
	resumos := make([]ruaResumo, 0, limite)
	for _, r := range ruas[:limite] {
		resumo := ruaResumo{
			Nome: r.NomeRua, Distrito: r.Distrito, Bairro: r.Bairro, CEP: r.CEP,
			TemGeometria: r.Geometria != "",
		}

		// Conhecimento de campo registrado por admin (rua sem saída,
		// numeração fora de ordem, etc.) — é isso que diferencia o Zé Rota
		// de só ler o cadastro cru.
		if s.obsRepo != nil {
			observacoes, err := s.obsRepo.FindByRuaID(ctx, r.ID)
			if err == nil {
				for _, o := range observacoes {
					resumo.Observacoes = append(resumo.Observacoes, fmt.Sprintf("[%s] %s", o.Categoria, o.Texto))
				}
			}
		}

		resumos = append(resumos, resumo)
	}

	saida, _ := json.Marshal(map[string]interface{}{
		"total_encontrado": len(ruas),
		"mostrando":        limite,
		"ruas":             resumos,
	})
	return string(saida)
}

// Coordenadas fixas do CDD Campos dos Goytacazes — suficiente pra previsão
// do tempo da cidade toda, não precisa de geolocalização por rua.
const (
	latCampos = -21.7545
	lonCampos = -41.3244
)

// consultarClima usa a Open-Meteo: API pública, gratuita, sem chave e sem
// limite prático pro nosso volume de uso. Devolve os dados brutos da API e
// deixa o próprio modelo interpretar e resumir pro carteiro (a instrução de
// como ler weather_code está no prompt do sistema).
func (s *zeRotaService) consultarClima(ctx context.Context) string {
	endpoint := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&current=temperature_2m,precipitation,rain,weather_code,wind_speed_10m&hourly=precipitation_probability,rain&forecast_days=1&timezone=America%%2FSao_Paulo",
		latCampos, lonCampos,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "erro: não consegui montar a consulta de clima"
	}

	resp, err := s.http.Do(req)
	if err != nil {
		return "erro: serviço de clima indisponível agora, tenta de novo em instantes"
	}
	defer resp.Body.Close()

	corpo, err := io.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != http.StatusOK {
		return "erro: serviço de clima respondeu de forma inesperada"
	}

	saida, _ := json.Marshal(map[string]interface{}{
		"fonte":        "Open-Meteo (gratuito, sem necessidade de chave)",
		"dados_brutos": json.RawMessage(corpo),
	})
	return string(saida)
}

// sugerirLinkMapa NÃO chama nenhuma API paga — é só montar uma URL pública
// de busca do Google Maps (google.com/maps/search), sem chave, sem limite,
// sem billing. Só serve de pista externa, nunca substitui o cadastro do CDD.
func (s *zeRotaService) sugerirLinkMapa(chamada groqToolCall) string {
	var entrada struct {
		Consulta string `json:"consulta"`
	}
	if err := json.Unmarshal([]byte(chamada.Function.Arguments), &entrada); err != nil || entrada.Consulta == "" {
		return "erro: não consegui ler o que buscar no mapa"
	}

	consultaCompleta := entrada.Consulta + ", Campos dos Goytacazes, RJ"
	link := "https://www.google.com/maps/search/?api=1&query=" + url.QueryEscape(consultaCompleta)

	saida, _ := json.Marshal(map[string]string{
		"aviso":            "isso é só uma sugestão de busca externa, NÃO é dado conferido no cadastro oficial do CDD — deixe isso claro pro carteiro",
		"link_google_maps": link,
	})
	return string(saida)
}
