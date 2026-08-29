package repositories

import (
	"context"

	"github.com/empresa/rotas-entrega/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DistritoRepository interface {
	FindAll(ctx context.Context) ([]models.Distrito, error)
	FindByCodigo(ctx context.Context, codigo string) (*models.Distrito, error)
	Upsert(ctx context.Context, distrito *models.Distrito) error
}

type distritoRepository struct {
	db *gorm.DB
}

func NewDistritoRepository(db *gorm.DB) DistritoRepository {
	return &distritoRepository{db: db}
}

// FindAll só traz distritos ativos — um distrito extinto por um
// redistritamento aplicado não deve aparecer no mapa nem nos chips de
// seleção pra ninguém. Pra fins internos (ex: reativar no futuro), use uma
// consulta direta; não há hoje uma tela que precise ver os inativos.
func (r *distritoRepository) FindAll(ctx context.Context) ([]models.Distrito, error) {
	var distritos []models.Distrito
	err := r.db.WithContext(ctx).Where("ativo = true").Order("codigo asc").Find(&distritos).Error
	return distritos, err
}

func (r *distritoRepository) FindByCodigo(ctx context.Context, codigo string) (*models.Distrito, error) {
	var distrito models.Distrito
	err := r.db.WithContext(ctx).First(&distrito, "codigo = ?", codigo).Error
	if err != nil {
		return nil, err
	}
	return &distrito, nil
}

// Upsert cria o distrito se não existir, ou atualiza nome/cor/geojson se já existir.
// Usa ON CONFLICT de verdade — o Save() do GORM não serve aqui porque a chave
// primária (codigo) nunca é zero-value, então ele tentaria só um UPDATE e
// silenciosamente não criaria nada se o registro ainda não existisse.
func (r *distritoRepository) Upsert(ctx context.Context, distrito *models.Distrito) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "codigo"}},
			DoUpdates: clause.AssignmentColumns([]string{"nome", "cor", "geojson", "updated_at"}),
		}).
		Create(distrito).Error
}
