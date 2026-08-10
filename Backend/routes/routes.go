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
//   - Público (sem token): mapa/distritos, busca de ruas, CEP, aniversariantes, health, swagger
//   - Autenticado (qualquer papel): listar/ver colaboradores, histórico, estatísticas
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
		// Consulta de ruas/rotas continua livre pra qualquer pessoa (mapa, busca, impressão)
		api.Get("/", ruaHandler.ListRuas)
		api.Get("/:id", ruaHandler.GetRua)
		// Escrita exige admin
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
		// Aniversariante do dia continua público — é a homenagem, sem dado sensível de lista completa
		colaboradores.Get("/aniversariantes-hoje", colaboradorHandler.AniversariantesHoje)
		// Resto da área de colaboradores exige login (qualquer papel pra ver, admin pra escrever)
		colaboradores.Get("/", autenticado, colaboradorHandler.ListColaboradores)
		colaboradores.Get("/:id", autenticado, colaboradorHandler.GetColaborador)
		colaboradores.Post("/", somenteAdmin, colaboradorHandler.CreateColaborador)
		colaboradores.Delete("/:id", somenteAdmin, colaboradorHandler.DeleteColaborador)
	}

	// Injeção de dependências - Folgas
	// Consulta por matrícula é pública (não exige login), mas precisa da
	// matrícula exata — não lista colaboradores. Lançar/excluir é admin-only.
	folgaRepo := repositories.NewFolgaRepository(db)
	folgaService := services.NewFolgaService(folgaRepo, colaboradorRepo, historicoRepo)
	folgaHandler := handlers.NewFolgaHandler(folgaService)

	folgas := app.Group("/folgas")
	{
		folgas.Get("/saldo", folgaHandler.ConsultarSaldo)
		folgas.Post("/", somenteAdmin, folgaHandler.LancarFolga)
		folgas.Delete("/:id", somenteAdmin, folgaHandler.ExcluirLancamento)
	}

	// Injeção de dependências - Observações de rua (conhecimento de campo
	// dos carteiros; leitura pública, escrita/exclusão só admin)
	ruaObsRepo := repositories.NewRuaObservacaoRepository(db)
	ruaObsService := services.NewRuaObservacaoService(ruaObsRepo, ruaRepo)
	ruaObsHandler := handlers.NewRuaObservacaoHandler(ruaObsService)
	app.Get("/ruas/:id/observacoes", ruaObsHandler.Listar)
	app.Post("/ruas/:id/observacoes", somenteAdmin, ruaObsHandler.Adicionar)
	app.Delete("/observacoes/:id", somenteAdmin, ruaObsHandler.Excluir)

	// Injeção de dependências - Zé Rota (chat público; sem chave de API
	// configurada, a rota responde erro amigável mas não derruba o resto)
	zeRotaService := services.NewZeRotaService(zeRotaWorkerURL, ruaRepo, ruaObsRepo)
	zeRotaHandler := handlers.NewZeRotaHandler(zeRotaService)
	app.Post("/ze-rota/conversar", zeRotaHandler.Conversar)

	// Injeção de dependências - Distritos (mapa é público)
	distritoRepo := repositories.NewDistritoRepository(db)
	distritoService := services.NewDistritoService(distritoRepo)
	distritoHandler := handlers.NewDistritoHandler(distritoService)

	distritos := app.Group("/distritos")
	{
		distritos.Get("/", distritoHandler.ListDistritos)
		distritos.Get("/:codigo", distritoHandler.GetDistrito)
	}

	// Injeção de dependências - Estatísticas
	// Fica público de propósito: é a faixa "Operação em números" que já aparece
	// embaixo do mapa pra qualquer visitante, e só expõe totais agregados (não
	// nome/matrícula/aniversário de ninguém). Se quiser restringir também, é só
	// trocar pra `autenticado` aqui.
	estatisticasHandler := handlers.NewEstatisticasHandler(ruaRepo, colaboradorRepo)
	app.Get("/estatisticas/operacao", estatisticasHandler.OperacaoEmNumeros)
}
