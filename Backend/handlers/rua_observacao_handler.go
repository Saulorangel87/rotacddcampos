package handlers

import (
	"strconv"

	"github.com/empresa/rotas-entrega/services"
	"github.com/gofiber/fiber/v2"
)

type RuaObservacaoHandler struct {
	service services.RuaObservacaoService
}

func NewRuaObservacaoHandler(service services.RuaObservacaoService) *RuaObservacaoHandler {
	return &RuaObservacaoHandler{service: service}
}

// Listar godoc
// @Summary Lista as observações de campo de uma rua
// @Description Rota pública — mesmo espírito da consulta de rua.
// @Tags rua-observacoes
// @Produce json
// @Param id path int true "ID da rua"
// @Success 200 {array} models.RuaObservacao
// @Router /ruas/{id}/observacoes [get]
func (h *RuaObservacaoHandler) Listar(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	obs, err := h.service.Listar(c.Context(), uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(obs)
}

// Adicionar godoc
// @Summary Adiciona uma observação de campo numa rua (admin)
// @Tags rua-observacoes
// @Accept json
// @Produce json
// @Param id path int true "ID da rua"
// @Param observacao body services.NovaObservacaoDTO true "Categoria e texto"
// @Success 201 {object} models.RuaObservacao
// @Failure 400 {object} map[string]interface{}
// @Router /ruas/{id}/observacoes [post]
func (h *RuaObservacaoHandler) Adicionar(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	var dto services.NovaObservacaoDTO
	if err := c.BodyParser(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo da requisição inválido"})
	}

	criadoPor, _ := c.Locals("matricula").(string)

	obs, err := h.service.Adicionar(c.Context(), uint(id), dto, criadoPor)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(obs)
}

// Excluir godoc
// @Summary Exclui uma observação de campo (admin)
// @Tags rua-observacoes
// @Param id path int true "ID da observação"
// @Success 204
// @Router /observacoes/{id} [delete]
func (h *RuaObservacaoHandler) Excluir(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	if err := h.service.Excluir(c.Context(), uint(id)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
