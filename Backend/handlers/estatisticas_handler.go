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
	Internos           int64 `json:"internos"`
	Administrativos    int64 `json:"administrativos"`
	OTT                int64 `json:"ott"`
	OT                 int64 `json:"ot"`
	Supervisores       int64 `json:"supervisores"`
	Gerentes           int64 `json:"gerentes"`
}

// OperacaoEmNumeros godoc
// @Summary Números gerais da operação (distritos, colaboradores, frota, funções administrativas)
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

	internos, err := h.colaboradorRepo.ContarPorFuncaoLike(ctx, "INTERNO")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	administrativos, err := h.colaboradorRepo.ContarPorFuncaoLike(ctx, "ADMINISTRATIVO")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// OTT e OT usam comparação exata (sem %), senão "OT" bateria dentro de "OTT" também.
	ott, err := h.colaboradorRepo.ContarPorFuncaoLike(ctx, "OTT")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	ot, err := h.colaboradorRepo.ContarPorFuncaoLike(ctx, "OT")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	supervisores, err := h.colaboradorRepo.ContarPorFuncaoLike(ctx, "SUPERVISOR")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	gerentes, err := h.colaboradorRepo.ContarPorFuncaoLike(ctx, "GERENTE")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(OperacaoResponse{
		TotalDistritos:     totalDistritos,
		TotalColaboradores: totalColaboradores,
		MotorizadosMoto:    moto,
		MotorizadosCarro:   carro,
		Ciclistas:          ciclistas,
		Internos:           internos,
		Administrativos:    administrativos,
		OTT:                ott,
		OT:                 ot,
		Supervisores:       supervisores,
		Gerentes:           gerentes,
	})
}
