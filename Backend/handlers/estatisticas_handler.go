package handlers

import (
	"github.com/empresa/rotas-entrega/repositories"
	"github.com/gofiber/fiber/v2"
)

type EstatisticasHandler struct {
	ruaRepo         repositories.RuaRepository
	colaboradorRepo repositories.ColaboradorRepository
}

func NewEstatisticasHandler(ruaRepo repositories.RuaRepository, colaboradorRepo repositories.ColaboradorRepository) *EstatisticasHandler {
	return &EstatisticasHandler{ruaRepo: ruaRepo, colaboradorRepo: colaboradorRepo}
}

type OperacaoResponse struct {
	TotalDistritos     int64 `json:"total_distritos"`
	TotalColaboradores int64 `json:"total_colaboradores"`
	MotorizadosMoto    int64 `json:"motorizados_moto"`
	MotorizadosCarro   int64 `json:"motorizados_carro"`
	Ciclistas          int64 `json:"ciclistas"`
}

// OperacaoEmNumeros godoc
// @Summary Números gerais da operação (distritos, colaboradores, frota)
// @Tags estatisticas
// @Produce json
// @Success 200 {object} OperacaoResponse
// @Router /estatisticas/operacao [get]
func (h *EstatisticasHandler) OperacaoEmNumeros(c *fiber.Ctx) error {
	ctx := c.Context()

	totalDistritos, err := h.ruaRepo.ContarDistritosDistintos(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	totalColaboradores, err := h.colaboradorRepo.ContarTotal(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	moto, err := h.colaboradorRepo.ContarPorFuncaoLike(ctx, "%motorizado (m%")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	carro, err := h.colaboradorRepo.ContarPorFuncaoLike(ctx, "%motorizado (v%")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	ciclistas, err := h.colaboradorRepo.ContarPorFuncaoLike(ctx, "%ciclista%")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(OperacaoResponse{
		TotalDistritos:     totalDistritos,
		TotalColaboradores: totalColaboradores,
		MotorizadosMoto:    moto,
		MotorizadosCarro:   carro,
		Ciclistas:          ciclistas,
	})
}