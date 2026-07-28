package middlewares

import "github.com/gofiber/fiber/v2"

// ErrorHandler padroniza respostas de erro da API.
func ErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "erro interno do servidor"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"error": message,
	})
}
