package handlers

import (
	"strconv"

	"github.com/empresa/rotas-entrega/services"
	"github.com/gofiber/fiber/v2"
)

type HistoricoHandler struct {
	service services.HistoricoService
}

func NewHistoricoHandler(service services.HistoricoService) *HistoricoHandler {
	return &HistoricoHandler{service: service}
}

// ListHistorico godoc
// @Summary Histórico de alterações de distrito
// @Description Lista paginada de todas as vezes que uma rua mudou de distrito
// @Tags historico
// @Produce json
// @Param pagina query int false "Página (padrão 1)"
// @Param limite query int false "Itens por página (padrão 10, máx 100)"
// @Success 200 {object} map[string]interface{}
// @Router /historico [get]
func (h *HistoricoHandler) ListHistorico(c *fiber.Ctx) error {
	pagina, _ := strconv.Atoi(c.Query("pagina", "1"))
	limite, _ := strconv.Atoi(c.Query("limite", "10"))

	registros, total, err := h.service.List(c.Context(), pagina, limite)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if pagina < 1 {
		pagina = 1
	}
	if limite < 1 || limite > 100 {
		limite = 10
	}
	totalPaginas := (total + int64(limite) - 1) / int64(limite)
	if totalPaginas < 1 {
		totalPaginas = 1
	}

	return c.JSON(fiber.Map{
		"itens":         registros,
		"total":         total,
		"pagina":        pagina,
		"total_paginas": totalPaginas,
	})
}
