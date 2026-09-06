package services

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/empresa/rotas-entrega/models"
	"gorm.io/gorm"
)

// Uma referência só é reutilizada para o mesmo conjunto de cadastros e a mesma
// origem. Mudar esta versão permite revisar a compatibilidade em evoluções futuras.
func contextoMemoriaOrdenamento() string {
	partida := PontoPartidaCDD()
	return fmt.Sprintf("cdd-campos:v1:%.6f:%.6f", partida.Latitude, partida.Longitude)
}

func assinaturaParadas(paradas []models.ParadaOrdenamento) (string, bool) {
	chaves := make([]string, 0, len(paradas))
	identificadas := len(paradas) > 0
	for _, parada := range paradas {
		chaves = append(chaves, parada.ChaveAgrupamento)
		id, err := strconv.ParseUint(strings.TrimPrefix(parada.ChaveAgrupamento, "rua:"), 10, 64)
		if !strings.HasPrefix(parada.ChaveAgrupamento, "rua:") || err != nil || id == 0 {
			identificadas = false
		}
	}
	sort.Strings(chaves)
	dados, _ := json.Marshal(chaves)
	return fmt.Sprintf("%x", sha256.Sum256(dados)), identificadas
}

// Mantém somente uma permutação completa; não deduz a posição de ruas novas,
// nem aplica uma lista parcial a outro percurso.
func aplicarSequencia(paradas []models.ParadaOrdenamento, sequencia []models.ParadaCorrecao, fonte string) bool {
	if len(paradas) == 0 || len(paradas) != len(sequencia) {
		return false
	}
	posicoes := make(map[string]int, len(sequencia))
	usadas := make(map[int]bool, len(sequencia))
	for _, item := range sequencia {
		if item.ChaveAgrupamento == "" || posicoes[item.ChaveAgrupamento] != 0 || item.OrdemFinal < 1 || item.OrdemFinal > len(sequencia) || usadas[item.OrdemFinal] {
			return false
		}
		posicoes[item.ChaveAgrupamento] = item.OrdemFinal
		usadas[item.OrdemFinal] = true
	}
	for _, parada := range paradas {
		if posicoes[parada.ChaveAgrupamento] == 0 {
			return false
		}
	}
	for indice := range paradas {
		ordem := posicoes[paradas[indice].ChaveAgrupamento]
		paradas[indice].OrdemFinal = &ordem
		paradas[indice].FonteOrdemFinal = fonte
	}
	return true
}

func preservarOrdemDaCarga(paradas, anteriores []models.ParadaOrdenamento) bool {
	sequencia := make([]models.ParadaCorrecao, 0, len(anteriores))
	fonte := "manual"
	for _, parada := range anteriores {
		if parada.OrdemFinal == nil {
			return false
		}
		sequencia = append(sequencia, models.ParadaCorrecao{ChaveAgrupamento: parada.ChaveAgrupamento, OrdemFinal: *parada.OrdemFinal})
		if parada.FonteOrdemFinal != "" {
			fonte = parada.FonteOrdemFinal
		}
	}
	return aplicarSequencia(paradas, sequencia, fonte)
}

func (s *ordenamentoService) EsquecerReferencia(ctx context.Context, usuarioID, referenciaID uint) error {
	if s.paradas == nil {
		return ErrReferenciaNaoEncontrada
	}
	err := s.paradas.DesativarReferencia(ctx, usuarioID, referenciaID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrReferenciaNaoEncontrada
	}
	return err
}
