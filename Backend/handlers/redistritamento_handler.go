package handlers

import (
	"strconv"

	"github.com/empresa/rotas-entrega/services"
	"github.com/gofiber/fiber/v2"
)

type RedistritamentoHandler struct {
	service services.RedistritamentoService
}

func NewRedistritamentoHandler(service services.RedistritamentoService) *RedistritamentoHandler {
	return &RedistritamentoHandler{service: service}
}

// GetPlanoAtivo godoc
// @Summary Buscar o plano de redistritamento em andamento (rascunho ou concluído)
// @Tags redistritamento
// @Produce json
// @Success 200 {object} models.PlanoRedistritamento
// @Router /redistritamento/ativo [get]
func (h *RedistritamentoHandler) GetPlanoAtivo(c *fiber.Ctx) error {
	plano, err := h.service.GetPlanoAtivo(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if plano == nil {
		return c.JSON(nil)
	}
	return c.JSON(plano)
}

type criarPlanoDTO struct {
	QuantidadeAlvo int `json:"quantidade_alvo"`
}

// CriarPlano godoc
// @Summary Iniciar um novo plano de redução de distritos
// @Tags redistritamento
// @Accept json
// @Produce json
// @Param body body criarPlanoDTO true "Quantidade alvo de distritos"
// @Success 201 {object} models.PlanoRedistritamento
// @Router /redistritamento/planos [post]
func (h *RedistritamentoHandler) CriarPlano(c *fiber.Ctx) error {
	var dto criarPlanoDTO
	if err := c.BodyParser(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo da requisição inválido"})
	}

	matricula, _ := c.Locals("matricula").(string)
	plano, err := h.service.CriarPlano(c.Context(), dto.QuantidadeAlvo, matricula)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(plano)
}

type reatribuirRuaDTO struct {
	DistritoDestino string `json:"distrito_destino"`
}

// ReatribuirRua godoc
// @Summary Definir o distrito de destino de uma rua órfã dentro de um plano
// @Tags redistritamento
// @Accept json
// @Param planoId path int true "ID do plano"
// @Param ruaId path int true "ID da linha plano_redistritamento_ruas (não é o id da rua)"
// @Param body body reatribuirRuaDTO true "Distrito de destino"
// @Success 200 {object} map[string]interface{}
// @Router /redistritamento/planos/{planoId}/ruas/{ruaId} [put]
func (h *RedistritamentoHandler) ReatribuirRua(c *fiber.Ctx) error {
	planoID, err := strconv.ParseUint(c.Params("planoId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id de plano inválido"})
	}
	planoRuaID, err := strconv.ParseUint(c.Params("ruaId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id de rua inválido"})
	}

	var dto reatribuirRuaDTO
	if err := c.BodyParser(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo da requisição inválido"})
	}

	if err := h.service.ReatribuirRua(c.Context(), uint(planoID), uint(planoRuaID), dto.DistritoDestino); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"ok": true})
}

// Concluir godoc
// @Summary Marcar o plano como concluído (ainda editável, mas pronto pra revisão)
// @Tags redistritamento
// @Produce json
// @Param planoId path int true "ID do plano"
// @Success 200 {object} models.PlanoRedistritamento
// @Router /redistritamento/planos/{planoId}/concluir [post]
func (h *RedistritamentoHandler) Concluir(c *fiber.Ctx) error {
	planoID, err := strconv.ParseUint(c.Params("planoId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id de plano inválido"})
	}

	plano, err := h.service.Concluir(c.Context(), uint(planoID))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(plano)
}

// Aplicar godoc
// @Summary Aplicar o plano de forma DEFINITIVA — move as ruas de verdade e desativa os distritos extintos
// @Tags redistritamento
// @Produce json
// @Param planoId path int true "ID do plano"
// @Success 200 {object} map[string]interface{}
// @Router /redistritamento/planos/{planoId}/aplicar [post]
func (h *RedistritamentoHandler) Aplicar(c *fiber.Ctx) error {
	planoID, err := strconv.ParseUint(c.Params("planoId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id de plano inválido"})
	}

	matricula, _ := c.Locals("matricula").(string)
	if err := h.service.Aplicar(c.Context(), uint(planoID), matricula); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"ok": true})
}

// Cancelar godoc
// @Summary Descartar um plano em rascunho/concluído e voltar pra tela inicial
// @Tags redistritamento
// @Produce json
// @Param planoId path int true "ID do plano"
// @Success 200 {object} map[string]interface{}
// @Router /redistritamento/planos/{planoId} [delete]
func (h *RedistritamentoHandler) Cancelar(c *fiber.Ctx) error {
	planoID, err := strconv.ParseUint(c.Params("planoId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id de plano inválido"})
	}

	if err := h.service.CancelarPlano(c.Context(), uint(planoID)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"ok": true})
}
