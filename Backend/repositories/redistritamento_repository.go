package repositories

import (
	"context"

	"github.com/empresa/rotas-entrega/models"
	"gorm.io/gorm"
)

type RedistritamentoRepository interface {
	// Plano ativo é o mais recente com status "rascunho" ou "concluido" — só
	// pode existir um por vez (regra de negócio validada no service).
	FindPlanoAtivo(ctx context.Context) (*models.PlanoRedistritamento, error)
	FindPlanoByID(ctx context.Context, id uint) (*models.PlanoRedistritamento, error)
	CreatePlano(ctx context.Context, plano *models.PlanoRedistritamento) error
	UpdatePlano(ctx context.Context, plano *models.PlanoRedistritamento) error

	UpdateDestinoRua(ctx context.Context, planoRuaID uint, distritoDestino string) error
	CountRuasSemDestino(ctx context.Context, planoID uint) (int64, error)

	// Aplicar roda tudo dentro de uma transação: move as ruas de verdade,
	// desativa os distritos extintos e grava o histórico de auditoria.
	Aplicar(ctx context.Context, plano *models.PlanoRedistritamento, matricula string) error

	// DeletePlano remove um plano em rascunho/concluído e suas linhas —
	// só é chamado com plano ainda não aplicado, então não afeta nada real.
	DeletePlano(ctx context.Context, planoID uint) error
}

type redistritamentoRepository struct {
	db *gorm.DB
}

func NewRedistritamentoRepository(db *gorm.DB) RedistritamentoRepository {
	return &redistritamentoRepository{db: db}
}

func (r *redistritamentoRepository) FindPlanoAtivo(ctx context.Context) (*models.PlanoRedistritamento, error) {
	var plano models.PlanoRedistritamento
	err := r.db.WithContext(ctx).
		Preload("Ruas").
		Where("status IN ?", []string{models.StatusRedistritamentoRascunho, models.StatusRedistritamentoConcluido}).
		Order("created_at desc").
		First(&plano).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &plano, nil
}

func (r *redistritamentoRepository) FindPlanoByID(ctx context.Context, id uint) (*models.PlanoRedistritamento, error) {
	var plano models.PlanoRedistritamento
	err := r.db.WithContext(ctx).Preload("Ruas").First(&plano, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &plano, nil
}

func (r *redistritamentoRepository) CreatePlano(ctx context.Context, plano *models.PlanoRedistritamento) error {
	return r.db.WithContext(ctx).Create(plano).Error
}

func (r *redistritamentoRepository) UpdatePlano(ctx context.Context, plano *models.PlanoRedistritamento) error {
	return r.db.WithContext(ctx).Save(plano).Error
}

func (r *redistritamentoRepository) UpdateDestinoRua(ctx context.Context, planoRuaID uint, distritoDestino string) error {
	return r.db.WithContext(ctx).
		Model(&models.PlanoRedistritamentoRua{}).
		Where("id = ?", planoRuaID).
		Update("distrito_destino", distritoDestino).Error
}

func (r *redistritamentoRepository) CountRuasSemDestino(ctx context.Context, planoID uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&models.PlanoRedistritamentoRua{}).
		Where("plano_id = ? AND (distrito_destino IS NULL OR distrito_destino = '')", planoID).
		Count(&total).Error
	return total, err
}

func (r *redistritamentoRepository) Aplicar(ctx context.Context, plano *models.PlanoRedistritamento, matricula string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, linha := range plano.Ruas {
			if err := tx.Model(&models.Rua{}).
				Where("id = ?", linha.RuaID).
				Update("distrito", linha.DistritoDestino).Error; err != nil {
				return err
			}

			historico := &models.HistoricoAlteracao{
				Tipo:            "rua",
				RuaID:           linha.RuaID,
				NomeRua:         linha.NomeRua,
				DistritoOrigem:  linha.DistritoOrigem,
				DistritoDestino: linha.DistritoDestino,
				Descricao:       "Redistritamento — plano de redução aplicado",
				Usuario:         matricula,
			}
			if err := tx.Create(historico).Error; err != nil {
				return err
			}
		}

		if plano.DistritosExtintos != "" {
			codigos := splitCSV(plano.DistritosExtintos)
			if len(codigos) > 0 {
				if err := tx.Model(&models.Distrito{}).
					Where("codigo IN ?", codigos).
					Update("ativo", false).Error; err != nil {
					return err
				}
			}
		}

		return tx.Model(&models.PlanoRedistritamento{}).
			Where("id = ?", plano.ID).
			Updates(map[string]interface{}{
				"status":       models.StatusRedistritamentoAplicado,
				"aplicado_em":  plano.AplicadoEm,
				"aplicado_por": matricula,
			}).Error
	})
}

func (r *redistritamentoRepository) DeletePlano(ctx context.Context, planoID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("plano_id = ?", planoID).Delete(&models.PlanoRedistritamentoRua{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.PlanoRedistritamento{}, "id = ?", planoID).Error
	})
}

func splitCSV(s string) []string {
	var out []string
	atual := ""
	for _, ch := range s {
		if ch == ',' {
			if atual != "" {
				out = append(out, atual)
				atual = ""
			}
			continue
		}
		atual += string(ch)
	}
	if atual != "" {
		out = append(out, atual)
	}
	return out
}
