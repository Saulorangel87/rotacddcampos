package handlers_test

import (
	"context"
	"net/http/httptest"
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
}

func (s *ordenamentoServiceFake) GetAtivo(_ context.Context, usuarioID uint) (*models.Ordenamento, error) {
	s.usuarioIDRecebido = usuarioID
	return nil, nil
}

func (s *ordenamentoServiceFake) Criar(_ context.Context, usuarioID uint) (*models.Ordenamento, error) {
	s.usuarioIDRecebido = usuarioID
	return &models.Ordenamento{ID: 1, UsuarioID: usuarioID, Status: models.StatusOrdenamentoEmAndamento}, nil
}

func TestOrdenamentoExigeAutenticacao(t *testing.T) {
	app, _ := novoAppOrdenamentoTeste(t)
	req := httptest.NewRequest("GET", "/ordenamentos/ativo", nil)

	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("requisição falhou: %v", err)
	}
	if res.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status = %d, esperado %d", res.StatusCode, fiber.StatusUnauthorized)
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
			{metodo: "POST", caminho: "/ordenamentos", statusEsperado: fiber.StatusCreated},
		} {
			t.Run(papel+"_"+caso.metodo, func(t *testing.T) {
				app, service := novoAppOrdenamentoTeste(t)
				req := httptest.NewRequest(caso.metodo, caso.caminho, nil)
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

func novoAppOrdenamentoTeste(t *testing.T) (*fiber.App, *ordenamentoServiceFake) {
	t.Helper()
	service := &ordenamentoServiceFake{}
	handler := handlers.NewOrdenamentoHandler(service)
	app := fiber.New()
	app.Get("/ordenamentos/ativo", middlewares.ExigirAutenticacao("segredo-teste"), handler.GetAtivo)
	app.Post("/ordenamentos", middlewares.ExigirAutenticacao("segredo-teste"), handler.Criar)
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
