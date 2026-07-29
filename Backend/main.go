package main

import (
	"log/slog"
	"os"

	"github.com/empresa/rotas-entrega/config"
	"github.com/empresa/rotas-entrega/database"
	"github.com/empresa/rotas-entrega/routes"
	_"github.com/empresa/rotas-entrega/docs"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// @title API Rotas de Entrega
// @version 1.0
// @description Sistema de consulta e gestão de ruas para rotas de entrega.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email suporte@empresa.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /
func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		slog.Error("falha ao conectar no banco", "error", err)
		os.Exit(1)
	}

	if err := database.RunMigrations(db); err != nil {
		slog.Error("falha ao executar migrations", "error", err)
		os.Exit(1)
	}

	app := fiber.New(fiber.Config{
		AppName:      "Rotas de Entrega API",
		ErrorHandler: customErrorHandler,
	})

	app.Use(recover.New())
	app.Use(logger.New())

	app.Use(cors.New(cors.Config{
    AllowOrigins: "http://localhost:5173",
    AllowHeaders: "Origin, Content-Type, Accept",
    AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
}))

	routes.SetupRoutes(app, db)

	slog.Info("servidor iniciado", "port", cfg.ServerPort)
	if err := app.Listen(":" + cfg.ServerPort); err != nil {
		slog.Error("falha ao iniciar servidor", "error", err)
		os.Exit(1)
	}
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}
