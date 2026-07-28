package handlers

import (
	"strconv"

	"github.com/empresa/rotas-entrega/services"
	"github.com/gofiber/fiber/v2"
)

type RuaHandler struct {
	service services.RuaService
}

func NewRuaHandler(service services.RuaService) *RuaHandler {
	return &RuaHandler{service: service}
}

// ListRuas godoc
// @Summary Listar ruas
// @Description Retorna lista de ruas com filtros opcionais por nome, CEP ou distrito
// @Tags ruas
// @Accept json
// @Produce json
// @Param nome query string false "Filtrar por nome da rua (parcial, case-insensitive)"
// @Param cep query string false "Filtrar por CEP (parcial)"
// @Param distrito query string false "Filtrar por distrito (parcial, case-insensitive)"
// @Success 200 {array} models.Rua
// @Failure 500 {object} map[string]interface{}
// @Router /ruas [get]
func (h *RuaHandler) ListRuas(c *fiber.Ctx) error {
	nome := c.Query("nome")
	cep := c.Query("cep")
	distrito := c.Query("distrito")

	ruas, err := h.service.List(c.Context(), nome, cep, distrito)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(ruas)
}

// GetRua godoc
// @Summary Buscar rua por ID
// @Description Retorna uma rua específica pelo ID
// @Tags ruas
// @Accept json
// @Produce json
// @Param id path int true "ID da Rua"
// @Success 200 {object} models.Rua
// @Failure 404 {object} map[string]interface{}
// @Router /ruas/{id} [get]
func (h *RuaHandler) GetRua(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	rua, err := h.service.GetByID(c.Context(), uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "rua não encontrada"})
	}
	return c.JSON(rua)
}

// CreateRua godoc
// @Summary Cadastrar rua
// @Description Cria um novo registro de rua
// @Tags ruas
// @Accept json
// @Produce json
// @Param rua body services.CreateRuaDTO true "Dados da rua"
// @Success 201 {object} models.Rua
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /ruas [post]
func (h *RuaHandler) CreateRua(c *fiber.Ctx) error {
	var dto services.CreateRuaDTO
	if err := c.BodyParser(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	rua, err := h.service.Create(c.Context(), dto)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(rua)
}

// UpdateRua godoc
// @Summary Atualizar rua
// @Description Atualiza os dados de uma rua existente
// @Tags ruas
// @Accept json
// @Produce json
// @Param id path int true "ID da Rua"
// @Param rua body services.UpdateRuaDTO true "Dados atualizados"
// @Success 200 {object} models.Rua
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /ruas/{id} [put]
func (h *RuaHandler) UpdateRua(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	var dto services.UpdateRuaDTO
	if err := c.BodyParser(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	rua, err := h.service.Update(c.Context(), uint(id), dto)
	if err != nil {
		if err.Error() == "rua não encontrada" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(rua)
}

// DeleteRua godoc
// @Summary Remover rua
// @Description Remove uma rua pelo ID
// @Tags ruas
// @Accept json
// @Produce json
// @Param id path int true "ID da Rua"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /ruas/{id} [delete]
func (h *RuaHandler) DeleteRua(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	if err := h.service.Delete(c.Context(), uint(id)); err != nil {
		if err.Error() == "rua não encontrada" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
