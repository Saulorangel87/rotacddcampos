package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/empresa/rotas-entrega/config"
	"github.com/empresa/rotas-entrega/database"
	"github.com/empresa/rotas-entrega/routes"
	_"github.com/empresa/rotas-entrega/docs"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
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
    AllowOrigins: cfg.CORSOrigins,
    AllowHeaders: "Origin, Content-Type, Accept, Authorization",
    AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
}))

	// Proteção extra contra força bruta: no máximo 10 tentativas de login por IP
	// a cada minuto. O bloqueio por conta (5 tentativas/15min) já existe no
	// service e continua valendo — isso aqui trava a varredura de matrículas
	// diferentes vindas do mesmo IP, o que o bloqueio por conta sozinho não pega.
	app.Use("/auth/login", limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "muitas tentativas de login vindas daqui, aguarde um minuto",
			})
		},
	}))

	routes.SetupRoutes(app, db, cfg.JWTSecret, cfg.JWTExpiracaoHoras, cfg.ZeRotaWorkerURL)

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
