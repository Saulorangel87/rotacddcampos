package handlers

import (
	"strconv"

	"github.com/empresa/rotas-entrega/services"
	"github.com/gofiber/fiber/v2"
)

type FolgaHandler struct {
	service services.FolgaService
}

func NewFolgaHandler(service services.FolgaService) *FolgaHandler {
	return &FolgaHandler{service: service}
}

// ConsultarSaldo godoc
// @Summary Consultar saldo de folgas por matrícula
// @Description Rota pública, mas exige a matrícula exata (não lista colaboradores). Retorna nome, saldo atual e extrato completo de créditos/débitos.
// @Tags folgas
// @Produce json
// @Param matricula query string true "Matrícula do colaborador"
// @Success 200 {object} services.SaldoFolgas
// @Failure 404 {object} map[string]interface{}
// @Router /folgas/saldo [get]
func (h *FolgaHandler) ConsultarSaldo(c *fiber.Ctx) error {
	matricula := c.Query("matricula")
	if matricula == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "informe a matrícula"})
	}

	saldo, err := h.service.ConsultarSaldo(c.Context(), matricula)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(saldo)
}

// LancarFolga godoc
// @Summary Lançar crédito ou débito de folga (admin)
// @Tags folgas
// @Accept json
// @Produce json
// @Param lancamento body services.NovoLancamentoDTO true "Dados do lançamento"
// @Success 201 {object} models.FolgaLancamento
// @Failure 400 {object} map[string]interface{}
// @Router /folgas [post]
func (h *FolgaHandler) LancarFolga(c *fiber.Ctx) error {
	var dto services.NovoLancamentoDTO
	if err := c.BodyParser(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo da requisição inválido"})
	}

	// Autor do lançamento vem do token, nunca do que o front manda — mesmo
	// cuidado já aplicado no histórico de movimentação de ruas.
	criadoPor, _ := c.Locals("matricula").(string)

	lancamento, err := h.service.Lancar(c.Context(), dto, criadoPor)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(lancamento)
}

// ExcluirLancamento godoc
// @Summary Excluir lançamento de folga (admin)
// @Tags folgas
// @Param id path int true "ID do lançamento"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Router /folgas/{id} [delete]
func (h *FolgaHandler) ExcluirLancamento(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	excluidoPor, _ := c.Locals("matricula").(string)

	if err := h.service.Excluir(c.Context(), uint(id), excluidoPor); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
