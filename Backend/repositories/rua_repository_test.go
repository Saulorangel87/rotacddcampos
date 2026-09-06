package repositories

import (
	"reflect"
	"testing"
)

func TestTermosBuscaRuaRemoveTiposEArtigos(t *testing.T) {
	testes := []struct {
		entrada  string
		esperado []string
	}{
		{entrada: "Av. Sete de Setembro", esperado: []string{"SETE", "SETEMBRO"}},
		{entrada: "Rua São Gonçalo", esperado: []string{"SAO", "GONCALO"}},
		{entrada: "Rua A", esperado: []string{"A"}},
	}

	for _, caso := range testes {
		if obtido := termosBuscaRua(caso.entrada); !reflect.DeepEqual(obtido, caso.esperado) {
			t.Fatalf("termosBuscaRua(%q) = %#v, esperado %#v", caso.entrada, obtido, caso.esperado)
		}
	}
}

func TestDigitosCEPRemoveFormatacao(t *testing.T) {
	if obtido := digitosCEP("CEP: 28010-562"); obtido != "28010562" {
		t.Fatalf("digitosCEP() = %q, esperado 28010562", obtido)
	}
}
