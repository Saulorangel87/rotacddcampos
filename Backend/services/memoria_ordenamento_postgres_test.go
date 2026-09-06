package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/empresa/rotas-entrega/config"
	"github.com/empresa/rotas-entrega/models"
	"github.com/empresa/rotas-entrega/repositories"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Opt-in: ORDENAMENTO_TESTE_POSTGRES=1 go test ./services -run TestMemoriaPostgres.
// Usa somente PostgreSQL local. Schema, fixtures e migrações ficam dentro de
// uma transação revertida ao fim; não acessa as tabelas operacionais.
func TestMemoriaPostgres(t *testing.T) {
	if os.Getenv("ORDENAMENTO_TESTE_POSTGRES") != "1" {
		t.Skip("teste de integração local opt-in")
	}
	valores, err := godotenv.Read("../.env")
	if err != nil {
		t.Fatal(err)
	}
	if valores["DB_HOST"] != "localhost" && valores["DB_HOST"] != "127.0.0.1" && valores["DB_HOST"] != "::1" {
		t.Fatal("teste exige DB_HOST local explícito")
	}
	cfg := config.Config{DBHost: valores["DB_HOST"], DBPort: valores["DB_PORT"], DBUser: valores["DB_USER"], DBPassword: valores["DB_PASSWORD"], DBName: valores["DB_NAME"], DBSSLMode: valores["DB_SSLMODE"]}
	if cfg.DBPort == "" {
		cfg.DBPort = "5432"
	}
	if cfg.DBSSLMode == "" {
		cfg.DBSSLMode = "disable"
	}
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	t.Cleanup(func() { tx.Rollback() })
	schema := fmt.Sprintf("ordem_teste_%d", time.Now().UnixNano())
	if err := tx.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Exec("SET LOCAL search_path TO " + schema).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.AutoMigrate(&models.Usuario{}, &models.Ordenamento{}, &models.ObjetoOrdenamento{}, &models.ParadaOrdenamento{}, &models.CorrecaoOrdenamento{}); err != nil {
		t.Fatal(err)
	}
	// Também confirma que a migração é repetível.
	if err := tx.AutoMigrate(&models.CorrecaoOrdenamento{}, &models.ParadaOrdenamento{}); err != nil {
		t.Fatal(err)
	}
	usuario := models.Usuario{Matricula: "teste-memoria", SenhaHash: "fixture", Papel: models.PapelColaborador}
	outro := models.Usuario{Matricula: "teste-outro", SenhaHash: "fixture", Papel: models.PapelAdmin}
	if err := tx.Create(&usuario).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Create(&outro).Error; err != nil {
		t.Fatal(err)
	}
	repo := repositories.NewOrdenamentoRepository(tx)
	paradas := repositories.NewParadaOrdenamentoRepository(tx)
	s := NewOrdenamentoService(repo, enderecoResolverFake{}, nil, paradas, otimizadorFake{})
	d, err := s.Criar(ctx, usuario.ID)
	if err != nil {
		t.Fatal(err)
	}
	ordenamentoID := d.ID
	adicionar := func() {
		lat, lon := -21.75, -41.32
		for i, nome := range []string{"RUA A", "RUA B"} {
			objeto := models.ObjetoOrdenamento{OrdenamentoID: ordenamentoID, NomeRua: nome, TextoEntrada: nome, ChaveAgrupamento: fmt.Sprintf("rua:%d", i+10), OrigemEntrada: models.OrigemEntradaManual, StatusResolucao: models.StatusResolucaoIdentificado, Latitude: &lat, Longitude: &lon}
			if err := repo.CreateObjeto(ctx, &objeto); err != nil {
				t.Fatal(err)
			}
		}
	}
	adicionar()
	d, err = s.GerarOrdem(ctx, usuario.ID, ordenamentoID, GerarOrdemDTO{})
	if err != nil {
		t.Fatal(err)
	}
	ids := []uint{d.OrdemSugerida[1].ID, d.OrdemSugerida[0].ID}
	salvo, err := s.SalvarOrdemFinal(ctx, usuario.ID, ordenamentoID, SalvarOrdemFinalDTO{ParadaIDs: ids, Reutilizar: true})
	if err != nil {
		t.Fatal(err)
	}
	if salvo.ReferenciaPessoalID == nil || len(salvo.HistoricoOrdens) != 1 || len(salvo.HistoricoOrdens[0].Paradas) != 2 {
		t.Fatal("histórico JSON não persistiu")
	}
	refID := *salvo.ReferenciaPessoalID
	if err := s.EsquecerReferencia(ctx, outro.ID, refID); !errors.Is(err, ErrReferenciaNaoEncontrada) {
		t.Fatal("outro usuário desativou referência", err)
	}
	if h, err := paradas.ListHistorico(ctx, outro.ID); err != nil || len(h) != 0 {
		t.Fatal("vazamento de histórico", err)
	}

	// Falha no INSERT do histórico deve reverter tanto a ordem quanto a
	// desativação da referência anterior (uma única transação).
	invalida := salvo.HistoricoOrdens[0]
	// Repetir a chave primária faz o INSERT falhar depois da desativação.
	if err := paradas.UpdateOrdemFinal(ctx, ordenamentoID, []uint{ids[1], ids[0]}, &invalida); err == nil {
		t.Fatal("histórico inválido deveria falhar")
	}
	d, err = s.GetAtivo(ctx, usuario.ID)
	if err != nil || !reflect.DeepEqual(sequenciaFinal(d), sequenciaFinal(salvo)) || len(d.HistoricoOrdens) != 1 || d.ReferenciaPessoalID == nil || *d.ReferenciaPessoalID != refID {
		t.Fatal("falha deixou gravação parcial", err)
	}
	invalida = salvo.HistoricoOrdens[0]
	invalida.ID = 0
	if err := paradas.UpdateOrdemFinal(ctx, ordenamentoID, ids[:1], &invalida); !errors.Is(err, repositories.ErrParadasAlteradas) {
		t.Fatal("aceitou lista incompleta", err)
	}

	// A limpeza real remove objetos/paradas, mas a memória sobrevive e usa
	// os novos IDs da carga seguinte.
	d, err = s.Limpar(ctx, usuario.ID, ordenamentoID)
	if err != nil || len(d.OrdemSugerida) != 0 || len(d.HistoricoOrdens) != 1 {
		t.Fatal("limpeza apagou histórico", err)
	}
	adicionar()
	d, err = s.GerarOrdem(ctx, usuario.ID, ordenamentoID, GerarOrdemDTO{})
	if err != nil || !reflect.DeepEqual(sequenciaFinal(d), sequenciaFinal(salvo)) {
		t.Fatal("não reaplicou memória persistida", err)
	}
	if d.OrdemSugerida[0].FonteOrdemFinal != "pessoal" {
		t.Fatal("origem da memória não foi persistida")
	}
	if _, err := s.SalvarOrdemFinal(ctx, usuario.ID, ordenamentoID, SalvarOrdemFinalDTO{ParadaIDs: ids}); !errors.Is(err, ErrOrdemFinalInvalida) {
		t.Fatal("aceitou IDs da carga antiga", err)
	}
	if err := s.EsquecerReferencia(ctx, usuario.ID, refID); err != nil {
		t.Fatal(err)
	}
	d, err = s.GetAtivo(ctx, usuario.ID)
	if err != nil || d.ReferenciaPessoalID != nil || len(d.HistoricoOrdens) != 1 || d.HistoricoOrdens[0].DesativadaEm == nil {
		t.Fatal("desativação inconsistente", err)
	}
}
