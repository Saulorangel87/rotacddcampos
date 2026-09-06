package handlers_test

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/empresa/rotas-entrega/handlers"
	"github.com/empresa/rotas-entrega/middlewares"
	"github.com/empresa/rotas-entrega/models"
	"github.com/empresa/rotas-entrega/services"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type ordenamentoServiceFake struct {
	usuarioIDRecebido uint
	salvarDTO         services.SalvarOrdemFinalDTO
	gerarDTO          services.GerarOrdemDTO
}

func (s *ordenamentoServiceFake) EsquecerReferencia(_ context.Context, usuarioID, _ uint) error {
	s.usuarioIDRecebido = usuarioID
	return nil
}

func (s *ordenamentoServiceFake) GetAtivo(_ context.Context, usuarioID uint) (*services.OrdenamentoDetalhe, error) {
	s.usuarioIDRecebido = usuarioID
	return nil, nil
}

func (s *ordenamentoServiceFake) Criar(_ context.Context, usuarioID uint) (*services.OrdenamentoDetalhe, error) {
	s.usuarioIDRecebido = usuarioID
	return &services.OrdenamentoDetalhe{ID: 1, Status: models.StatusOrdenamentoEmAndamento}, nil
}

func (s *ordenamentoServiceFake) AdicionarObjeto(_ context.Context, usuarioID, _ uint, _ services.AdicionarObjetoDTO) (*services.OrdenamentoDetalhe, error) {
	s.usuarioIDRecebido = usuarioID
	return &services.OrdenamentoDetalhe{ID: 1}, nil
}

func (s *ordenamentoServiceFake) SelecionarRua(_ context.Context, usuarioID, _, _, _ uint) (*services.OrdenamentoDetalhe, error) {
	s.usuarioIDRecebido = usuarioID
	return &services.OrdenamentoDetalhe{ID: 1}, nil
}

func (s *ordenamentoServiceFake) ExcluirObjeto(_ context.Context, usuarioID, _, _ uint) (*services.OrdenamentoDetalhe, error) {
	s.usuarioIDRecebido = usuarioID
	return &services.OrdenamentoDetalhe{ID: 1}, nil
}

func (s *ordenamentoServiceFake) Limpar(_ context.Context, usuarioID, _ uint) (*services.OrdenamentoDetalhe, error) {
	s.usuarioIDRecebido = usuarioID
	return &services.OrdenamentoDetalhe{ID: 1}, nil
}

func (s *ordenamentoServiceFake) GerarOrdem(_ context.Context, usuarioID, _ uint, dto services.GerarOrdemDTO) (*services.OrdenamentoDetalhe, error) {
	s.usuarioIDRecebido = usuarioID
	s.gerarDTO = dto
	return &services.OrdenamentoDetalhe{ID: 1}, nil
}

func (s *ordenamentoServiceFake) SalvarOrdemFinal(_ context.Context, usuarioID, _ uint, dto services.SalvarOrdemFinalDTO) (*services.OrdenamentoDetalhe, error) {
	s.usuarioIDRecebido = usuarioID
	s.salvarDTO = dto
	return &services.OrdenamentoDetalhe{ID: 1}, nil
}

func TestOrdenamentoExigeAutenticacao(t *testing.T) {
	for _, caso := range []struct {
		metodo  string
		caminho string
		corpo   string
	}{
		{metodo: "GET", caminho: "/ordenamentos/ativo"},
		{metodo: "DELETE", caminho: "/ordenamentos/referencias/1"},
		{metodo: "POST", caminho: "/ordenamentos"},
		{metodo: "POST", caminho: "/ordenamentos/1/objetos", corpo: `{"entrada":"Pelinca 520"}`},
		{metodo: "PATCH", caminho: "/ordenamentos/1/objetos/2/rua", corpo: `{"rua_id":10}`},
		{metodo: "DELETE", caminho: "/ordenamentos/1/objetos/2"},
		{metodo: "DELETE", caminho: "/ordenamentos/1/objetos"},
		{metodo: "POST", caminho: "/ordenamentos/1/gerar-ordem"},
		{metodo: "PATCH", caminho: "/ordenamentos/1/ordem-final", corpo: `{"parada_ids":[2,1]}`},
	} {
		t.Run(caso.metodo+caso.caminho, func(t *testing.T) {
			app, _ := novoAppOrdenamentoTeste(t)
			req := httptest.NewRequest(caso.metodo, caso.caminho, strings.NewReader(caso.corpo))

			res, err := app.Test(req)
			if err != nil {
				t.Fatalf("requisição falhou: %v", err)
			}
			if res.StatusCode != fiber.StatusUnauthorized {
				t.Fatalf("status = %d, esperado %d", res.StatusCode, fiber.StatusUnauthorized)
			}
		})
	}
}

func TestOrdenamentoPermiteColaboradorEAdmin(t *testing.T) {
	for _, papel := range []string{models.PapelColaborador, models.PapelAdmin} {
		for _, caso := range []struct {
			metodo         string
			caminho        string
			statusEsperado int
		}{
			{metodo: "GET", caminho: "/ordenamentos/ativo", statusEsperado: fiber.StatusOK},
			{metodo: "DELETE", caminho: "/ordenamentos/referencias/1", statusEsperado: fiber.StatusNoContent},
			{metodo: "POST", caminho: "/ordenamentos", statusEsperado: fiber.StatusCreated},
			{metodo: "POST", caminho: "/ordenamentos/1/objetos", statusEsperado: fiber.StatusCreated},
			{metodo: "PATCH", caminho: "/ordenamentos/1/objetos/2/rua", statusEsperado: fiber.StatusOK},
			{metodo: "DELETE", caminho: "/ordenamentos/1/objetos/2", statusEsperado: fiber.StatusOK},
			{metodo: "DELETE", caminho: "/ordenamentos/1/objetos", statusEsperado: fiber.StatusOK},
			{metodo: "POST", caminho: "/ordenamentos/1/gerar-ordem", statusEsperado: fiber.StatusOK},
			{metodo: "PATCH", caminho: "/ordenamentos/1/ordem-final", statusEsperado: fiber.StatusOK},
		} {
			t.Run(papel+"_"+caso.metodo, func(t *testing.T) {
				app, service := novoAppOrdenamentoTeste(t)
				req := httptest.NewRequest(caso.metodo, caso.caminho, nil)
				if caso.metodo == "POST" && strings.HasSuffix(caso.caminho, "/objetos") {
					req = httptest.NewRequest(caso.metodo, caso.caminho, strings.NewReader(`{"entrada":"Pelinca 520"}`))
					req.Header.Set("Content-Type", "application/json")
				}
				if caso.metodo == "PATCH" {
					corpo := `{"rua_id":10}`
					if strings.HasSuffix(caso.caminho, "/ordem-final") {
						corpo = `{"parada_ids":[2,1]}`
					}
					req = httptest.NewRequest(caso.metodo, caso.caminho, strings.NewReader(corpo))
					req.Header.Set("Content-Type", "application/json")
				}
				req.Header.Set("Authorization", "Bearer "+tokenTeste(t, papel))

				res, err := app.Test(req)
				if err != nil {
					t.Fatalf("requisição falhou: %v", err)
				}
				if res.StatusCode != caso.statusEsperado {
					t.Fatalf("status = %d, esperado %d", res.StatusCode, caso.statusEsperado)
				}
				if service.usuarioIDRecebido != 42 {
					t.Fatalf("usuarioID = %d, esperado 42", service.usuarioIDRecebido)
				}
			})
		}
	}
}

func TestOrdenamentoOpcoesDaMemoriaUsamIdentidadeDoJWT(t *testing.T) {
	app, service := novoAppOrdenamentoTeste(t)
	for _, caso := range []struct{ metodo, caminho, corpo string }{
		{"PATCH", "/ordenamentos/1/ordem-final", `{"parada_ids":[2,1],"reutilizar":true,"usuario_id":99}`},
		{"POST", "/ordenamentos/1/gerar-ordem", `{"somente_algoritmo":true,"usuario_id":99}`},
	} {
		req := httptest.NewRequest(caso.metodo, caso.caminho, strings.NewReader(caso.corpo))
		req.Header.Set("Authorization", "Bearer "+tokenTeste(t, models.PapelColaborador))
		req.Header.Set("Content-Type", "application/json")
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != fiber.StatusOK || service.usuarioIDRecebido != 42 {
			t.Fatal("identidade deve vir do JWT")
		}
	}
	if !service.salvarDTO.Reutilizar || len(service.salvarDTO.ParadaIDs) != 2 || service.salvarDTO.ParadaIDs[0] != 2 || !service.gerarDTO.SomenteAlgoritmo {
		t.Fatal("API não repassou as opções da memória")
	}
}

func novoAppOrdenamentoTeste(t *testing.T) (*fiber.App, *ordenamentoServiceFake) {
	t.Helper()
	service := &ordenamentoServiceFake{}
	handler := handlers.NewOrdenamentoHandler(service)
	app := fiber.New()
	app.Get("/ordenamentos/ativo", middlewares.ExigirAutenticacao("segredo-teste"), handler.GetAtivo)
	app.Delete("/ordenamentos/referencias/:referenciaId", middlewares.ExigirAutenticacao("segredo-teste"), handler.EsquecerReferencia)
	app.Post("/ordenamentos", middlewares.ExigirAutenticacao("segredo-teste"), handler.Criar)
	app.Post("/ordenamentos/:id/objetos", middlewares.ExigirAutenticacao("segredo-teste"), handler.AdicionarObjeto)
	app.Patch("/ordenamentos/:id/objetos/:objetoId/rua", middlewares.ExigirAutenticacao("segredo-teste"), handler.SelecionarRua)
	app.Delete("/ordenamentos/:id/objetos", middlewares.ExigirAutenticacao("segredo-teste"), handler.Limpar)
	app.Delete("/ordenamentos/:id/objetos/:objetoId", middlewares.ExigirAutenticacao("segredo-teste"), handler.ExcluirObjeto)
	app.Post("/ordenamentos/:id/gerar-ordem", middlewares.ExigirAutenticacao("segredo-teste"), handler.GerarOrdem)
	app.Patch("/ordenamentos/:id/ordem-final", middlewares.ExigirAutenticacao("segredo-teste"), handler.SalvarOrdemFinal)
	return app, service
}

func tokenTeste(t *testing.T, papel string) string {
	t.Helper()
	claims := services.Claims{
		UsuarioID: 42,
		Matricula: "teste",
		Papel:     papel,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("segredo-teste"))
	if err != nil {
		t.Fatalf("não foi possível assinar token de teste: %v", err)
	}
	return token
}
