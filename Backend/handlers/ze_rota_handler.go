package handlers

import (
	"errors"

	"github.com/empresa/rotas-entrega/services"
	"github.com/gofiber/fiber/v2"
)

type ZeRotaHandler struct {
	service services.ZeRotaService
}

func NewZeRotaHandler(service services.ZeRotaService) *ZeRotaHandler {
	return &ZeRotaHandler{service: service}
}

type conversarRequestDTO struct {
	Mensagens []services.MensagemChat `json:"mensagens"`
}

// Conversar godoc
// @Summary Conversa com o Zé Rota (chat de rotas)
// @Description Rota pública. Recebe o histórico completo da conversa (o front guarda o estado, o backend não persiste nada) e devolve a próxima resposta do Zé Rota.
// @Tags ze-rota
// @Accept json
// @Produce json
// @Param corpo body conversarRequestDTO true "Histórico da conversa"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /ze-rota/conversar [post]
func (h *ZeRotaHandler) Conversar(c *fiber.Ctx) error {
	var dto conversarRequestDTO
	if err := c.BodyParser(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo da requisição inválido"})
	}

	resposta, err := h.service.Conversar(c.Context(), dto.Mensagens)
	if err != nil {
		// Falha de infraestrutura (Worker/Groq fora do ar, rede) é 502, não
		// 400 — 400 fica reservado pra requisição realmente malformada. O
		// erro real (com detalhe técnico) já foi logado no servidor pelo
		// serviço; aqui só repassamos a mensagem amigável.
		if errors.Is(err, services.ErrFalhaUpstreamZeRota) {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"resposta": resposta})
}
