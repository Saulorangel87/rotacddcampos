package handlers

import (
	"errors"
	"strconv"

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

func (h *OrdenamentoHandler) AdicionarObjeto(c *fiber.Ctx) error {
	usuarioID, ok := c.Locals("usuario_id").(uint)
	if !ok || usuarioID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "não autenticado"})
	}
	ordenamentoID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || ordenamentoID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id de ordenamento inválido"})
	}

	var dto services.AdicionarObjetoDTO
	if err := c.BodyParser(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo da requisição inválido"})
	}
	detalhe, err := h.service.AdicionarObjeto(c.Context(), usuarioID, uint(ordenamentoID), dto)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEntradaVazia):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, services.ErrOrdenamentoNaoEncontrado):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "não foi possível adicionar a encomenda"})
		}
	}
	return c.Status(fiber.StatusCreated).JSON(detalhe)
}

func (h *OrdenamentoHandler) SelecionarRua(c *fiber.Ctx) error {
	usuarioID, ok := c.Locals("usuario_id").(uint)
	if !ok || usuarioID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "não autenticado"})
	}
	ordenamentoID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || ordenamentoID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id de ordenamento inválido"})
	}
	objetoID, err := strconv.ParseUint(c.Params("objetoId"), 10, 64)
	if err != nil || objetoID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id de encomenda inválido"})
	}
	var corpo struct {
		RuaID uint `json:"rua_id"`
	}
	if err := c.BodyParser(&corpo); err != nil || corpo.RuaID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "rua inválida"})
	}
	detalhe, err := h.service.SelecionarRua(c.Context(), usuarioID, uint(ordenamentoID), uint(objetoID), corpo.RuaID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrOrdenamentoNaoEncontrado), errors.Is(err, services.ErrObjetoNaoEncontrado):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, services.ErrOpcaoRuaInvalida):
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "não foi possível confirmar a rua"})
		}
	}
	return c.JSON(detalhe)
}

func (h *OrdenamentoHandler) ExcluirObjeto(c *fiber.Ctx) error {
	usuarioID, ok := c.Locals("usuario_id").(uint)
	if !ok || usuarioID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "não autenticado"})
	}
	ordenamentoID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || ordenamentoID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id de ordenamento inválido"})
	}
	objetoID, err := strconv.ParseUint(c.Params("objetoId"), 10, 64)
	if err != nil || objetoID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id de encomenda inválido"})
	}

	detalhe, err := h.service.ExcluirObjeto(c.Context(), usuarioID, uint(ordenamentoID), uint(objetoID))
	if err != nil {
		if errors.Is(err, services.ErrOrdenamentoNaoEncontrado) || errors.Is(err, services.ErrObjetoNaoEncontrado) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "não foi possível remover a encomenda"})
	}
	return c.JSON(detalhe)
}

// Limpar remove todos os objetos e a ordem sugerida do ordenamento ativo.
func (h *OrdenamentoHandler) Limpar(c *fiber.Ctx) error {
	usuarioID, ok := c.Locals("usuario_id").(uint)
	if !ok || usuarioID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "não autenticado"})
	}
	ordenamentoID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || ordenamentoID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id de ordenamento inválido"})
	}

	detalhe, err := h.service.Limpar(c.Context(), usuarioID, uint(ordenamentoID))
	if err != nil {
		if errors.Is(err, services.ErrOrdenamentoNaoEncontrado) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "não foi possível limpar a lista"})
	}
	return c.JSON(detalhe)
}

// GerarOrdem monta a sequência sugerida de ruas para o ordenamento ativo.
func (h *OrdenamentoHandler) GerarOrdem(c *fiber.Ctx) error {
	usuarioID, ok := c.Locals("usuario_id").(uint)
	if !ok || usuarioID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "não autenticado"})
	}
	ordenamentoID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || ordenamentoID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id de ordenamento inválido"})
	}

	detalhe, err := h.service.GerarOrdem(c.Context(), usuarioID, uint(ordenamentoID))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrOrdenamentoNaoEncontrado):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, services.ErrOrdenamentoSemObjetos),
			errors.Is(err, services.ErrOrdenamentoComPendencias),
			errors.Is(err, services.ErrOrdenamentoSemCoordenadas):
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "não foi possível gerar o ordenamento"})
		}
	}
	return c.JSON(detalhe)
}
