package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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
Fala como um carteiro experiente e prestativo, em português brasileiro, direto e sem enrolação.

Regra mais importante: você NUNCA inventa nome de rua, distrito, CEP ou qualquer dado de endereço de
memória — mesmo que pareça óbvio ou você "ache" que sabe. Toda vez que a pergunta envolver onde fica
uma rua, qual o distrito dela, ou qualquer informação de endereço, você usa a ferramenta buscar_rua para
confirmar antes de responder. Se a busca não encontrar nada, diga isso claramente e sugira conferir a
grafia do nome — não tente adivinhar.

Quando a busca trouxer "observacoes_de_campo", são anotações reais de carteiros sobre aquela rua
(acesso difícil, numeração fora de ordem, mais de um nome, questão de segurança, etc.) — sempre mencione
isso na resposta, é informação valiosa pra quem não conhece a rua.

Se a pergunta não tiver nada a ver com ruas, distritos ou entregas de Campos dos Goytacazes, explique
educadamente que só pode ajudar com isso.

Respostas curtas e práticas — quem está perguntando é um carteiro no meio da rota, não alguém lendo um
relatório.`

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
}

func texto(s string) *string { return &s }

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
			return *escolha.Message.Content, nil
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
	if chamada.Function.Name != "buscar_rua" {
		return "erro: ferramenta desconhecida"
	}

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
