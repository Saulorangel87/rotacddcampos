package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/empresa/rotas-entrega/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrParadasAlteradas = errors.New("a lista mudou; atualize a página e confira a sequência antes de salvar")

type ParadaOrdenamentoRepository interface {
	ListByOrdenamento(ctx context.Context, ordenamentoID uint) ([]models.ParadaOrdenamento, error)
	ReplaceByOrdenamento(ctx context.Context, ordenamentoID uint, paradas []models.ParadaOrdenamento) error
	UpdateOrdemFinal(ctx context.Context, ordenamentoID uint, paradaIDs []uint, correcao *models.CorrecaoOrdenamento) error
	DeleteByOrdenamento(ctx context.Context, ordenamentoID uint) error
	FindReferencia(ctx context.Context, usuarioID uint, contexto, assinatura string) (*models.CorrecaoOrdenamento, error)
	ListHistorico(ctx context.Context, usuarioID uint) ([]models.CorrecaoOrdenamento, error)
	DesativarReferencia(ctx context.Context, usuarioID, referenciaID uint) error
}

type paradaOrdenamentoRepository struct {
	db *gorm.DB
}

func NewParadaOrdenamentoRepository(db *gorm.DB) ParadaOrdenamentoRepository {
	return &paradaOrdenamentoRepository{db: db}
}

func (r *paradaOrdenamentoRepository) ListByOrdenamento(ctx context.Context, ordenamentoID uint) ([]models.ParadaOrdenamento, error) {
	var paradas []models.ParadaOrdenamento
	err := r.db.WithContext(ctx).Where("ordenamento_id = ?", ordenamentoID).Order("ordem_sugerida ASC").Find(&paradas).Error
	return paradas, err
}

func (r *paradaOrdenamentoRepository) ReplaceByOrdenamento(ctx context.Context, ordenamentoID uint, paradas []models.ParadaOrdenamento) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("ordenamento_id = ?", ordenamentoID).Delete(&models.ParadaOrdenamento{}).Error; err != nil {
			return err
		}
		if len(paradas) == 0 {
			return nil
		}
		return tx.Create(&paradas).Error
	})
}

func (r *paradaOrdenamentoRepository) UpdateOrdemFinal(ctx context.Context, ordenamentoID uint, paradaIDs []uint, correcao *models.CorrecaoOrdenamento) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Serializa publicação/desativação da memória pessoal deste usuário.
		var usuario models.Usuario
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&usuario, correcao.UsuarioID).Error; err != nil {
			return err
		}
		var ordenamento models.Ordenamento
		if err := tx.Where("id = ? AND usuario_id = ? AND status = ?", ordenamentoID, correcao.UsuarioID, models.StatusOrdenamentoEmAndamento).First(&ordenamento).Error; err != nil {
			return err
		}
		var atuais []models.ParadaOrdenamento
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("ordenamento_id = ?", ordenamentoID).Order("id ASC").Find(&atuais).Error; err != nil {
			return err
		}
		ids := make(map[uint]bool, len(atuais))
		for _, parada := range atuais {
			ids[parada.ID] = true
		}
		if len(paradaIDs) == 0 || len(paradaIDs) != len(atuais) {
			return ErrParadasAlteradas
		}
		for _, id := range paradaIDs {
			if !ids[id] {
				return ErrParadasAlteradas
			}
			delete(ids, id)
		}

		for indice, paradaID := range paradaIDs {
			resultado := tx.Model(&models.ParadaOrdenamento{}).
				Where("ordenamento_id = ? AND id = ?", ordenamentoID, paradaID).
				Updates(map[string]any{"ordem_final": indice + 1, "fonte_ordem_final": "manual"})
			if resultado.Error != nil {
				return resultado.Error
			}
			if resultado.RowsAffected != 1 {
				return gorm.ErrRecordNotFound
			}
		}
		if correcao.Reutilizar {
			if err := tx.Model(&models.CorrecaoOrdenamento{}).
				Where("usuario_id = ? AND contexto = ? AND assinatura_ruas = ? AND reutilizar = ? AND desativada_em IS NULL", correcao.UsuarioID, correcao.Contexto, correcao.AssinaturaRuas, true).
				Update("desativada_em", time.Now()).Error; err != nil {
				return err
			}
		}
		return tx.Create(correcao).Error
	})
}

func (r *paradaOrdenamentoRepository) FindReferencia(ctx context.Context, usuarioID uint, contexto, assinatura string) (*models.CorrecaoOrdenamento, error) {
	var correcao models.CorrecaoOrdenamento
	err := r.db.WithContext(ctx).Where("usuario_id = ? AND contexto = ? AND assinatura_ruas = ? AND reutilizar = ? AND desativada_em IS NULL", usuarioID, contexto, assinatura, true).Order("id DESC").First(&correcao).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &correcao, err
}

func (r *paradaOrdenamentoRepository) ListHistorico(ctx context.Context, usuarioID uint) ([]models.CorrecaoOrdenamento, error) {
	correcoes := make([]models.CorrecaoOrdenamento, 0)
	err := r.db.WithContext(ctx).Where("usuario_id = ?", usuarioID).Order("id DESC").Limit(20).Find(&correcoes).Error
	return correcoes, err
}

func (r *paradaOrdenamentoRepository) DesativarReferencia(ctx context.Context, usuarioID, referenciaID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var usuario models.Usuario
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&usuario, usuarioID).Error; err != nil {
			return err
		}
		resultado := tx.Model(&models.CorrecaoOrdenamento{}).
			Where("id = ? AND usuario_id = ? AND reutilizar = ? AND desativada_em IS NULL", referenciaID, usuarioID, true).
			Update("desativada_em", time.Now())
		if resultado.Error != nil {
			return resultado.Error
		}
		if resultado.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (r *paradaOrdenamentoRepository) DeleteByOrdenamento(ctx context.Context, ordenamentoID uint) error {
	return r.db.WithContext(ctx).Where("ordenamento_id = ?", ordenamentoID).Delete(&models.ParadaOrdenamento{}).Error
}
