package handlers

import (
	"errors"
	"strconv"

	"github.com/empresa/rotas-entrega/services"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	service services.AuthService
}

func NewAuthHandler(service services.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Login godoc
// @Summary Login
// @Description Autentica por matrícula + senha e retorna um JWT. Bloqueia a conta por 15min após 5 tentativas erradas.
// @Tags auth
// @Accept json
// @Produce json
// @Param login body services.LoginDTO true "Credenciais"
// @Success 200 {object} services.LoginResponse
// @Failure 401 {object} map[string]interface{}
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var dto services.LoginDTO
	if err := c.BodyParser(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo da requisição inválido"})
	}

	resposta, err := h.service.Login(c.Context(), dto, c.IP())
	if err != nil {
		switch {
		case errors.Is(err, services.ErrContaBloqueada):
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.JSON(resposta)
}

// TrocarSenha godoc
// @Summary Trocar a própria senha
// @Description Exige token válido. Usado tanto pra troca voluntária quanto pra sair do estado de senha provisória.
// @Tags auth
// @Accept json
// @Produce json
// @Param dados body services.TrocarSenhaDTO true "Senha atual e nova"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Router /auth/trocar-senha [post]
func (h *AuthHandler) TrocarSenha(c *fiber.Ctx) error {
	usuarioID, ok := c.Locals("usuario_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "não autenticado"})
	}

	var dto services.TrocarSenhaDTO
	if err := c.BodyParser(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo da requisição inválido"})
	}

	if err := h.service.TrocarSenha(c.Context(), usuarioID, dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ListUsuarios godoc
// @Summary Listar usuários de acesso (admin)
// @Description Só admin pode chamar. Não retorna hash de senha nem dados sensíveis.
// @Tags auth
// @Produce json
// @Success 200 {array} services.UsuarioResumo
// @Router /auth/usuarios [get]
func (h *AuthHandler) ListUsuarios(c *fiber.Ctx) error {
	usuarios, err := h.service.ListUsuarios(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(usuarios)
}

// ResetarSenha godoc
// @Summary Resetar a senha de outro usuário (admin)
// @Description Só admin pode chamar. Não pede a senha atual — define uma nova senha temporária e marca como provisória.
// @Tags auth
// @Accept json
// @Produce json
// @Param id path int true "ID do usuário"
// @Param dados body services.ResetarSenhaDTO true "Nova senha temporária"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Router /auth/usuarios/{id}/resetar-senha [post]
func (h *AuthHandler) ResetarSenha(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	var dto services.ResetarSenhaDTO
	if err := c.BodyParser(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo da requisição inválido"})
	}

	if err := h.service.ResetarSenha(c.Context(), uint(id), dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
// @Summary Cadastrar novo usuário de acesso (admin)
// @Description Só admin pode chamar. Cria a conta com senha temporária — usuário troca no primeiro login.
// @Tags auth
// @Accept json
// @Produce json
// @Param dados body services.NovoUsuarioDTO true "Dados do novo usuário"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /auth/usuarios [post]
func (h *AuthHandler) CriarUsuario(c *fiber.Ctx) error {
	var dto services.NovoUsuarioDTO
	if err := c.BodyParser(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo da requisição inválido"})
	}

	usuario, err := h.service.CriarUsuario(c.Context(), dto)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":        usuario.ID,
		"matricula": usuario.Matricula,
		"papel":     usuario.Papel,
	})
}
