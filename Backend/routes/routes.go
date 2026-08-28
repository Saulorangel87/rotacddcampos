package routes

import (
	"github.com/empresa/rotas-entrega/handlers"
	"github.com/empresa/rotas-entrega/middlewares"
	"github.com/empresa/rotas-entrega/repositories"
	"github.com/empresa/rotas-entrega/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"gorm.io/gorm"
)

// SetupRoutes monta todas as rotas da API.
// jwtSecret e jwtHoras vêm da config (config.JWTSecret / config.JWTExpiracaoHoras).
//
// Modelo de acesso:
//   - Público (sem token): apenas health e swagger (infraestrutura)
//   - Autenticado (qualquer papel — colaborador ou admin): tudo o mais que é
//     leitura — mapa/distritos, busca de ruas, CEP, aniversariantes, folgas,
//     observações de rua, Zé Rota, colaboradores, histórico, estatísticas
//   - Admin: criar/editar/excluir ruas e colaboradores, gerenciar usuários
func SetupRoutes(app *fiber.App, db *gorm.DB, jwtSecret string, jwtHoras int, zeRotaWorkerURL string) {
	autenticado := middlewares.ExigirAutenticacao(jwtSecret)
	somenteAdmin := middlewares.ExigirAdmin(jwtSecret)

	// Swagger
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Injeção de dependências - Auth
	usuarioRepo := repositories.NewUsuarioRepository(db)
	acessoLoginRepo := repositories.NewAcessoLoginRepository(db)
	authService := services.NewAuthService(usuarioRepo, acessoLoginRepo, jwtSecret, jwtHoras)
	authHandler := handlers.NewAuthHandler(authService)

	auth := app.Group("/auth")
	{
		auth.Post("/login", authHandler.Login)
		auth.Post("/trocar-senha", autenticado, authHandler.TrocarSenha)
		auth.Get("/usuarios", somenteAdmin, authHandler.ListUsuarios)
		auth.Post("/usuarios", somenteAdmin, authHandler.CriarUsuario)
		auth.Post("/usuarios/:id/resetar-senha", somenteAdmin, authHandler.ResetarSenha)
	}

	// Injeção de dependências - Histórico (precisa existir antes do RuaService, que grava nela)
	historicoRepo := repositories.NewHistoricoRepository(db)
	historicoService := services.NewHistoricoService(historicoRepo)
	historicoHandler := handlers.NewHistoricoHandler(historicoService)
	app.Get("/historico", autenticado, historicoHandler.ListHistorico)

	// Injeção de dependências - Ruas
	ruaRepo := repositories.NewRuaRepository(db)
	ruaService := services.NewRuaService(ruaRepo, historicoRepo)
	ruaHandler := handlers.NewRuaHandler(ruaService)

	api := app.Group("/ruas")
	{
		// Consulta de ruas/rotas agora exige login (qualquer papel) — acesso
		// restrito a funcionário da empresa, não é mais público
		api.Get("/", autenticado, ruaHandler.ListRuas)
		api.Get("/:id", autenticado, ruaHandler.GetRua)
		// Escrita continua exigindo admin
		api.Post("/", somenteAdmin, ruaHandler.CreateRua)
		api.Put("/:id", somenteAdmin, ruaHandler.UpdateRua)
		api.Delete("/:id", somenteAdmin, ruaHandler.DeleteRua)
	}

	// Injeção de dependências - Colaboradores
	colaboradorRepo := repositories.NewColaboradorRepository(db)
	colaboradorService := services.NewColaboradorService(colaboradorRepo)
	colaboradorHandler := handlers.NewColaboradorHandler(colaboradorService)

	colaboradores := app.Group("/colaboradores")
	{
		// Acesso público removido — agora exige login como o resto da área de
		// colaboradores (qualquer papel pra ver, admin pra escrever)
		colaboradores.Get("/aniversariantes-hoje", autenticado, colaboradorHandler.AniversariantesHoje)
		colaboradores.Get("/", autenticado, colaboradorHandler.ListColaboradores)
		colaboradores.Get("/:id", autenticado, colaboradorHandler.GetColaborador)
		colaboradores.Post("/", somenteAdmin, colaboradorHandler.CreateColaborador)
		colaboradores.Delete("/:id", somenteAdmin, colaboradorHandler.DeleteColaborador)
	}

	// Injeção de dependências - Folgas
	// Consulta por matrícula agora exige login — antes era pública mesmo
	// precisando de matrícula exata. Lançar/excluir continua admin-only.
	folgaRepo := repositories.NewFolgaRepository(db)
	folgaService := services.NewFolgaService(folgaRepo, colaboradorRepo, historicoRepo)
	folgaHandler := handlers.NewFolgaHandler(folgaService)

	folgas := app.Group("/folgas")
	{
		folgas.Get("/saldo", autenticado, folgaHandler.ConsultarSaldo)
		folgas.Post("/", somenteAdmin, folgaHandler.LancarFolga)
		folgas.Delete("/:id", somenteAdmin, folgaHandler.ExcluirLancamento)
	}

	// Injeção de dependências - Observações de rua (conhecimento de campo
	// dos carteiros; agora exige login pra leitura também, escrita/exclusão só admin)
	ruaObsRepo := repositories.NewRuaObservacaoRepository(db)
	ruaObsService := services.NewRuaObservacaoService(ruaObsRepo, ruaRepo)
	ruaObsHandler := handlers.NewRuaObservacaoHandler(ruaObsService)
	app.Get("/ruas/:id/observacoes", autenticado, ruaObsHandler.Listar)
	app.Post("/ruas/:id/observacoes", somenteAdmin, ruaObsHandler.Adicionar)
	app.Delete("/observacoes/:id", somenteAdmin, ruaObsHandler.Excluir)

	// Injeção de dependências - Zé Rota (agora exige login — antes era chat público)
	zeRotaService := services.NewZeRotaService(zeRotaWorkerURL, ruaRepo, ruaObsRepo)
	zeRotaHandler := handlers.NewZeRotaHandler(zeRotaService)
	app.Post("/ze-rota/conversar", autenticado, zeRotaHandler.Conversar)

	// Injeção de dependências - Distritos (agora exige login — mapa não é mais público)
	distritoRepo := repositories.NewDistritoRepository(db)
	distritoService := services.NewDistritoService(distritoRepo)
	distritoHandler := handlers.NewDistritoHandler(distritoService)

	distritos := app.Group("/distritos")
	{
		distritos.Get("/", autenticado, distritoHandler.ListDistritos)
		distritos.Get("/:codigo", autenticado, distritoHandler.GetDistrito)
	}

	// Injeção de dependências - Estatísticas (agora exige login)
	estatisticasHandler := handlers.NewEstatisticasHandler(ruaRepo, colaboradorRepo)
	app.Get("/estatisticas/operacao", autenticado, estatisticasHandler.OperacaoEmNumeros)
}
