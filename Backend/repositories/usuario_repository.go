package repositories

import (
	"context"

	"github.com/empresa/rotas-entrega/models"
	"gorm.io/gorm"
)

type UsuarioRepository interface {
	FindByMatricula(ctx context.Context, matricula string) (*models.Usuario, error)
	FindByID(ctx context.Context, id uint) (*models.Usuario, error)
	FindAll(ctx context.Context) ([]models.Usuario, error)
	Create(ctx context.Context, usuario *models.Usuario) error
	Update(ctx context.Context, usuario *models.Usuario) error
}

type usuarioRepository struct {
	db *gorm.DB
}

func NewUsuarioRepository(db *gorm.DB) UsuarioRepository {
	return &usuarioRepository{db: db}
}

func (r *usuarioRepository) FindByMatricula(ctx context.Context, matricula string) (*models.Usuario, error) {
	var usuario models.Usuario
	err := r.db.WithContext(ctx).Where("matricula = ?", matricula).First(&usuario).Error
	if err != nil {
		return nil, err
	}
	return &usuario, nil
}

func (r *usuarioRepository) FindByID(ctx context.Context, id uint) (*models.Usuario, error) {
	var usuario models.Usuario
	err := r.db.WithContext(ctx).First(&usuario, id).Error
	if err != nil {
		return nil, err
	}
	return &usuario, nil
}

func (r *usuarioRepository) FindAll(ctx context.Context) ([]models.Usuario, error) {
	var usuarios []models.Usuario
	err := r.db.WithContext(ctx).Order("matricula").Find(&usuarios).Error
	return usuarios, err
}

func (r *usuarioRepository) Create(ctx context.Context, usuario *models.Usuario) error {
	return r.db.WithContext(ctx).Create(usuario).Error
}

func (r *usuarioRepository) Update(ctx context.Context, usuario *models.Usuario) error {
	return r.db.WithContext(ctx).Save(usuario).Error
}
