package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNominatimAceitaSomenteResultadoDeCampos(t *testing.T) {
	servidor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" || r.URL.Query().Get("city") != municipioGeocodificacao {
			t.Errorf("requisição inesperada: %s", r.URL.String())
		}
		if r.Header.Get("User-Agent") != "aplicacao-teste" {
			t.Errorf("User-Agent = %q", r.Header.Get("User-Agent"))
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[
			{"lat":"-22.90","lon":"-43.20","display_name":"Rua Teste, Rio de Janeiro","address":{"city":"Rio de Janeiro"}},
			{"lat":"-21.7545","lon":"-41.3240","display_name":"Rua Teste, Campos dos Goytacazes, RJ, Brasil","address":{"municipality":"Campos dos Goytacazes"}}
		]`)
	}))
	defer servidor.Close()

	geocodificador := NewNominatimGeocodificador(servidor.URL, "aplicacao-teste")
	resultado, err := geocodificador.BuscarRua(context.Background(), "RUA TESTE")
	if err != nil {
		t.Fatalf("BuscarRua() erro inesperado: %v", err)
	}
	if resultado == nil || resultado.Latitude != -21.7545 || resultado.Longitude != -41.3240 {
		t.Fatalf("resultado inesperado: %+v", resultado)
	}
}

func TestNominatimNaoEnviaNumeroDoObjeto(t *testing.T) {
	servidor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Query().Get("street"), "342") {
			t.Errorf("a consulta externa recebeu número do objeto")
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[]`)
	}))
	defer servidor.Close()

	geocodificador := NewNominatimGeocodificador(servidor.URL, "aplicacao-teste")
	if _, err := geocodificador.BuscarRua(context.Background(), "AVENIDA SETE DE SETEMBRO"); err != nil {
		t.Fatalf("BuscarRua() erro inesperado: %v", err)
	}
}
