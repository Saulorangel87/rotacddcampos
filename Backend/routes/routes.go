package routes

import (
	"github.com/empresa/rotas-entrega/handlers"
	"github.com/empresa/rotas-entrega/repositories"
	"github.com/empresa/rotas-entrega/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App, db *gorm.DB) {
	// Swagger
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Injeção de dependências
	ruaRepo := repositories.NewRuaRepository(db)
	ruaService := services.NewRuaService(ruaRepo)
	ruaHandler := handlers.NewRuaHandler(ruaService)

	api := app.Group("/ruas")
	{
		api.Get("/", ruaHandler.ListRuas)
		api.Get("/:id", ruaHandler.GetRua)
		api.Post("/", ruaHandler.CreateRua)
		api.Put("/:id", ruaHandler.UpdateRua)
		api.Delete("/:id", ruaHandler.DeleteRua)
	}
}
