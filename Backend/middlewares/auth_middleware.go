package middlewares

import (
	"strings"

	"github.com/empresa/rotas-entrega/models"
	"github.com/empresa/rotas-entrega/services"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// ExigirAutenticacao valida o JWT do header Authorization e injeta
// usuario_id, matricula e papel no contexto da requisição. Usado nas rotas
// que exigem qualquer usuário logado (colaboradores, aniversariantes p/ nível "colaborador").
func ExigirAutenticacao(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, err := parseToken(c, jwtSecret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "não autenticado"})
		}

		c.Locals("usuario_id", claims.UsuarioID)
		c.Locals("matricula", claims.Matricula)
		c.Locals("papel", claims.Papel)
		return c.Next()
	}
}

// ExigirAdmin faz o mesmo que ExigirAutenticacao e ainda barra quem não é admin.
// Sempre combine com ExigirAutenticacao antes (ou use direto — ele já valida o token sozinho).
func ExigirAdmin(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, err := parseToken(c, jwtSecret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "não autenticado"})
		}

		if claims.Papel != models.PapelAdmin {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "acesso restrito a administradores"})
		}

		c.Locals("usuario_id", claims.UsuarioID)
		c.Locals("matricula", claims.Matricula)
		c.Locals("papel", claims.Papel)
		return c.Next()
	}
}

func parseToken(c *fiber.Ctx, jwtSecret string) (*services.Claims, error) {
	header := c.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return nil, fiber.ErrUnauthorized
	}
	tokenStr := strings.TrimPrefix(header, "Bearer ")

	claims := &services.Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, fiber.ErrUnauthorized
	}
	return claims, nil
}
