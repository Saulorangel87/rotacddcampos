package handlers

import (
	"errors"

	"github.com/empresa/rotas-entrega/services"
	"github.com/gofiber/fiber/v2"
)

type OrdenamentoHandler struct {
	service services.OrdenamentoService
}

func NewOrdenamentoHandler(service services.OrdenamentoService) *OrdenamentoHandler {
	return &OrdenamentoHandler{service: service}
}

// GetAtivo devolve somente o ordenamento em andamento do usuário autenticado.
func (h *OrdenamentoHandler) GetAtivo(c *fiber.Ctx) error {
	usuarioID, ok := c.Locals("usuario_id").(uint)
	if !ok || usuarioID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "não autenticado"})
	}

	ordenamento, err := h.service.GetAtivo(c.Context(), usuarioID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "não foi possível consultar o ordenamento"})
	}
	return c.JSON(ordenamento)
}

// Criar inicia um ordenamento vazio para o usuário autenticado.
func (h *OrdenamentoHandler) Criar(c *fiber.Ctx) error {
	usuarioID, ok := c.Locals("usuario_id").(uint)
	if !ok || usuarioID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "não autenticado"})
	}

	ordenamento, err := h.service.Criar(c.Context(), usuarioID)
	if err != nil {
		if errors.Is(err, services.ErrOrdenamentoJaExiste) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "não foi possível iniciar o ordenamento"})
	}
	return c.Status(fiber.StatusCreated).JSON(ordenamento)
}
