package handlers

import (
	"github.com/empresa/rotas-entrega/services"
	"github.com/gofiber/fiber/v2"
)

type DistritoHandler struct {
	service services.DistritoService
}

func NewDistritoHandler(service services.DistritoService) *DistritoHandler {
	return &DistritoHandler{service: service}
}

// ListDistritos godoc
// @Summary Listar distritos
// @Description Retorna todos os distritos cadastrados (nome, cor, geojson quando disponível)
// @Tags distritos
// @Produce json
// @Success 200 {array} models.Distrito
// @Router /distritos [get]
func (h *DistritoHandler) ListDistritos(c *fiber.Ctx) error {
	distritos, err := h.service.List(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(distritos)
}

// GetDistrito godoc
// @Summary Buscar distrito por código
// @Tags distritos
// @Produce json
// @Param codigo path string true "Código do distrito (ex: 601)"
// @Success 200 {object} models.Distrito
// @Failure 404 {object} map[string]interface{}
// @Router /distritos/{codigo} [get]
func (h *DistritoHandler) GetDistrito(c *fiber.Ctx) error {
	codigo := c.Params("codigo")
	distrito, err := h.service.GetByCodigo(c.Context(), codigo)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "distrito não encontrado"})
	}
	return c.JSON(distrito)
}
