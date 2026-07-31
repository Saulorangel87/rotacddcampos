package handlers

import (
	"strconv"
	"strings"

	"github.com/empresa/rotas-entrega/models"
	"github.com/empresa/rotas-entrega/services"
	"github.com/gofiber/fiber/v2"
)

type ColaboradorHandler struct {
	service services.ColaboradorService
}

type ColaboradorResponse struct {
	ID             uint   `json:"id"`
	Matricula      string `json:"matricula"`
	Nome           string `json:"nome"`
	Funcao         string `json:"funcao"`
	DataAdmissao   string `json:"data_admissao"`
	DataNascimento string `json:"data_nascimento"`
	Cargo          string `json:"cargo"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

func toColaboradorResponse(col models.Colaborador) ColaboradorResponse {
	response := ColaboradorResponse{
		ID:        col.ID,
		Matricula: col.Matricula,
		Nome:      col.Nome,
		Funcao:    col.Funcao,
		Cargo:     col.Cargo,
		CreatedAt: col.CreatedAt.Format("02/01/2006 15:04:05"),
		UpdatedAt: col.UpdatedAt.Format("02/01/2006 15:04:05"),
	}

	if col.DataAdmissao != nil {
		response.DataAdmissao = col.DataAdmissao.Format("02/01/2006")
	}
	if col.DataNascimento != nil {
		response.DataNascimento = col.DataNascimento.Format("02/01/2006")
	}
	return response
}

func toColaboradoresResponse(colaboradores []models.Colaborador) []ColaboradorResponse {
	responses := make([]ColaboradorResponse, 0, len(colaboradores))
	for _, col := range colaboradores {
		responses = append(responses, toColaboradorResponse(col))
	}
	return responses
}

func NewColaboradorHandler(service services.ColaboradorService) *ColaboradorHandler {
	return &ColaboradorHandler{service: service}
}

// ListColaboradores godoc
// @Summary Listar colaboradores
// @Description Retorna lista de colaboradores com filtros opcionais por nome ou matrícula
// @Tags colaboradores
// @Produce json
// @Param nome query string false "Filtrar por nome (parcial, case-insensitive)"
// @Param matricula query string false "Filtrar por matrícula (parcial)"
// @Param carteiro query string false "Filtrar apenas carteiros (true para somente carteiros)"
// @Success 200 {array} models.Colaborador
// @Router /colaboradores [get]
func (h *ColaboradorHandler) ListColaboradores(c *fiber.Ctx) error {
	nome := c.Query("nome")
	matricula := c.Query("matricula")
	carteiro := c.Query("carteiro")

	colaboradores, err := h.service.List(c.Context(), nome, matricula, carteiro)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(toColaboradoresResponse(colaboradores))
}

// GetColaborador godoc
// @Summary Buscar colaborador por ID
// @Tags colaboradores
// @Produce json
// @Param id path int true "ID do colaborador"
// @Success 200 {object} models.Colaborador
// @Failure 404 {object} map[string]interface{}
// @Router /colaboradores/{id} [get]
func (h *ColaboradorHandler) GetColaborador(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	colaborador, err := h.service.GetByID(c.Context(), uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "colaborador não encontrado"})
	}
	return c.JSON(toColaboradorResponse(*colaborador))
}

// CreateColaborador godoc
// @Summary Cadastrar novo colaborador
// @Tags colaboradores
// @Accept json
// @Produce json
// @Param colaborador body services.NovoColaboradorDTO true "Dados do novo colaborador"
// @Success 201 {object} models.Colaborador
// @Failure 400 {object} map[string]interface{}
// @Router /colaboradores [post]
func (h *ColaboradorHandler) CreateColaborador(c *fiber.Ctx) error {
	var dto services.NovoColaboradorDTO
	if err := c.BodyParser(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo da requisição inválido"})
	}

	colaborador, err := h.service.Create(c.Context(), dto)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(toColaboradorResponse(*colaborador))
}

// AniversariantesHoje godoc
// @Summary Aniversariantes do dia
// @Description Retorna colaboradores que fazem aniversário hoje. Aceita ?data=DD/MM ou ?data=DD-MM para testar outra data.
// @Tags colaboradores
// @Produce json
// @Param data query string false "Data no formato DD/MM ou DD-MM para simular (ex: 25/12 ou 25-12)"
// @Success 200 {array} models.Colaborador
// @Router /colaboradores/aniversariantes-hoje [get]
func (h *ColaboradorHandler) AniversariantesHoje(c *fiber.Ctx) error {
	dataSimulada := c.Query("data")

	if dataSimulada != "" {
		dataSimulada = strings.ReplaceAll(dataSimulada, "/", "-")
		partes := strings.Split(dataSimulada, "-")
		if len(partes) != 2 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "formato esperado: DD/MM ou DD-MM"})
		}
		dia, errDia := strconv.Atoi(partes[0])
		mes, errMes := strconv.Atoi(partes[1])
		if errMes != nil || errDia != nil || mes < 1 || mes > 12 || dia < 1 || dia > 31 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "data inválida, use DD/MM ou DD-MM"})
		}

		colaboradores, err := h.service.AniversariantesDaData(c.Context(), mes, dia)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(toColaboradoresResponse(colaboradores))
	}

	colaboradores, err := h.service.AniversariantesDeHoje(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(toColaboradoresResponse(colaboradores))
}
