package repositories

import (
	"context"
	"time"

	"github.com/empresa/rotas-entrega/models"
	"gorm.io/gorm"
)

type OrdenamentoRepository interface {
	FindAtivoByUsuario(ctx context.Context, usuarioID uint) (*models.Ordenamento, error)
	Create(ctx context.Context, ordenamento *models.Ordenamento) error
	ListObjetos(ctx context.Context, ordenamentoID uint) ([]models.ObjetoOrdenamento, error)
	CreateObjeto(ctx context.Context, objeto *models.ObjetoOrdenamento) error
	UpdateObjeto(ctx context.Context, objeto *models.ObjetoOrdenamento) error
	DeleteObjeto(ctx context.Context, ordenamentoID, objetoID uint) (bool, error)
	LimparConteudo(ctx context.Context, ordenamentoID uint) error
}

type ordenamentoRepository struct {
	db *gorm.DB
}

func NewOrdenamentoRepository(db *gorm.DB) OrdenamentoRepository {
	return &ordenamentoRepository{db: db}
}

func (r *ordenamentoRepository) FindAtivoByUsuario(ctx context.Context, usuarioID uint) (*models.Ordenamento, error) {
	var ordenamento models.Ordenamento
	err := r.db.WithContext(ctx).
		Where("usuario_id = ? AND status = ?", usuarioID, models.StatusOrdenamentoEmAndamento).
		Order("created_at DESC").
		First(&ordenamento).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &ordenamento, nil
}

func (r *ordenamentoRepository) Create(ctx context.Context, ordenamento *models.Ordenamento) error {
	return r.db.WithContext(ctx).Create(ordenamento).Error
}

func (r *ordenamentoRepository) ListObjetos(ctx context.Context, ordenamentoID uint) ([]models.ObjetoOrdenamento, error) {
	var objetos []models.ObjetoOrdenamento
	err := r.db.WithContext(ctx).
		Where("ordenamento_id = ?", ordenamentoID).
		Order("created_at ASC, id ASC").
		Find(&objetos).Error
	return objetos, err
}

func (r *ordenamentoRepository) CreateObjeto(ctx context.Context, objeto *models.ObjetoOrdenamento) error {
	return r.db.WithContext(ctx).Create(objeto).Error
}

func (r *ordenamentoRepository) UpdateObjeto(ctx context.Context, objeto *models.ObjetoOrdenamento) error {
	return r.db.WithContext(ctx).Save(objeto).Error
}

func (r *ordenamentoRepository) DeleteObjeto(ctx context.Context, ordenamentoID, objetoID uint) (bool, error) {
	resultado := r.db.WithContext(ctx).
		Where("id = ? AND ordenamento_id = ?", objetoID, ordenamentoID).
		Delete(&models.ObjetoOrdenamento{})
	return resultado.RowsAffected > 0, resultado.Error
}

// LimparConteudo apaga os dados operacionais da carga atual e reinicia o
// horário do ordenamento. O registro continua ativo para receber a nova carga.
func (r *ordenamentoRepository) LimparConteudo(ctx context.Context, ordenamentoID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("ordenamento_id = ?", ordenamentoID).Delete(&models.ParadaOrdenamento{}).Error; err != nil {
			return err
		}
		if err := tx.Where("ordenamento_id = ?", ordenamentoID).Delete(&models.ObjetoOrdenamento{}).Error; err != nil {
			return err
		}
		agora := time.Now()
		return tx.Model(&models.Ordenamento{}).
			Where("id = ?", ordenamentoID).
			Updates(map[string]any{"created_at": agora, "updated_at": agora}).Error
	})
}
