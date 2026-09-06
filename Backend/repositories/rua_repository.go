package repositories

import (
	"context"
	"strings"
	"unicode"

	"github.com/empresa/rotas-entrega/models"
	"golang.org/x/text/unicode/norm"
	"gorm.io/gorm"
)

type RuaRepository interface {
	FindAll(ctx context.Context, filters map[string]string) ([]models.Rua, error)
	FindByID(ctx context.Context, id uint) (*models.Rua, error)
	Create(ctx context.Context, rua *models.Rua) error
	Update(ctx context.Context, rua *models.Rua) error
	Delete(ctx context.Context, id uint) error
	ContarDistritosDistintos(ctx context.Context) (int64, error)
	// FindByDistritosExato busca por código exato (não ILIKE parcial) — usado
	// pelo redistritamento, onde "617" não pode acidentalmente casar com "6170".
	FindByDistritosExato(ctx context.Context, codigos []string) ([]models.Rua, error)
}

type ruaRepository struct {
	db *gorm.DB
}

func NewRuaRepository(db *gorm.DB) RuaRepository {
	return &ruaRepository{db: db}
}

func (r *ruaRepository) FindAll(ctx context.Context, filters map[string]string) ([]models.Rua, error) {
	var ruas []models.Rua
	query := r.db.WithContext(ctx).Model(&models.Rua{})

	if nome, ok := filters["nome"]; ok && nome != "" {
		// Cada palavra relevante precisa aparecer no cadastro. Isso permite
		// buscas como "Av Sete Setembro" mesmo quando o cadastro traz
		// "AVENIDA SETE DE SETEMBRO", sem devolver apenas resultados que
		// coincidem com uma única palavra da entrada.
		for _, token := range termosBuscaRua(nome) {
			query = query.Where("unaccent(nome_rua) ILIKE unaccent(?)", "%"+token+"%")
		}
	}
	if cep, ok := filters["cep"]; ok && cep != "" {
		// O cadastro pode conter CEP com ou sem hífen. Comparar somente os
		// dígitos evita que a forma usada na etiqueta altere o resultado.
		if somenteDigitos := digitosCEP(cep); somenteDigitos != "" {
			query = query.Where("regexp_replace(cep, '[^0-9]', '', 'g') LIKE ?", "%"+somenteDigitos+"%")
		} else {
			query = query.Where("cep LIKE ?", "%"+cep+"%")
		}
	}
	if distrito, ok := filters["distrito"]; ok && distrito != "" {
		query = query.Where("distrito ILIKE ?", "%"+distrito+"%")
	}

	err := query.Order("nome_rua asc").Order("id asc").Find(&ruas).Error
	return ruas, err
}

var palavrasIgnoradasBuscaRua = map[string]bool{
	"A": true, "AS": true, "O": true, "OS": true,
	"DA": true, "DAS": true, "DE": true, "DO": true, "DOS": true, "E": true,
}

var tiposIgnoradosBuscaRua = map[string]bool{
	"R": true, "RUA": true, "AV": true, "AVENIDA": true,
	"TV": true, "TRAVESSA": true, "PCA": true, "PRACA": true,
	"EST": true, "ESTRADA": true, "RODOVIA": true, "ALAMEDA": true, "LARGO": true,
	"BOULEVARD": true,
}

func termosBuscaRua(valor string) []string {
	valor = strings.ToUpper(valor)
	valor = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return ' '
	}, valor)

	partes := strings.Fields(valor)
	semTipo := make([]string, 0, len(partes))
	for _, parte := range partes {
		// unaccent no PostgreSQL compara os acentos, mas as listas locais usam
		// as formas sem acento para reconhecer os tipos e artigos.
		parte = strings.Map(func(r rune) rune {
			if unicode.Is(unicode.Mn, r) {
				return -1
			}
			return r
		}, norm.NFD.String(parte))
		if tiposIgnoradosBuscaRua[parte] {
			continue
		}
		semTipo = append(semTipo, parte)
	}
	filtradas := make([]string, 0, len(semTipo))
	for _, parte := range semTipo {
		if !palavrasIgnoradasBuscaRua[parte] {
			filtradas = append(filtradas, parte)
		}
	}
	if len(filtradas) > 0 {
		return filtradas
	}
	// "Rua" ou "A" ainda devem filtrar, em vez de virar uma listagem
	// completa. Só removemos artigos quando restou pelo menos uma palavra útil;
	// se só houver um artigo, preservamos a palavra sem o tipo do logradouro.
	if len(semTipo) > 0 {
		return semTipo
	}
	return partes
}

func digitosCEP(valor string) string {
	var resultado strings.Builder
	for _, r := range valor {
		if unicode.IsDigit(r) {
			resultado.WriteRune(r)
		}
	}
	return resultado.String()
}

func (r *ruaRepository) FindByDistritosExato(ctx context.Context, codigos []string) ([]models.Rua, error) {
	var ruas []models.Rua
	err := r.db.WithContext(ctx).
		Where("distrito IN ? AND ativo = true", codigos).
		Order("nome_rua asc").
		Find(&ruas).Error
	return ruas, err
}

func (r *ruaRepository) FindByID(ctx context.Context, id uint) (*models.Rua, error) {
	var rua models.Rua
	err := r.db.WithContext(ctx).First(&rua, id).Error
	if err != nil {
		return nil, err
	}
	return &rua, nil
}

func (r *ruaRepository) Create(ctx context.Context, rua *models.Rua) error {
	return r.db.WithContext(ctx).Create(rua).Error
}

func (r *ruaRepository) Update(ctx context.Context, rua *models.Rua) error {
	return r.db.WithContext(ctx).Save(rua).Error
}

func (r *ruaRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Rua{}, id).Error
}

// ContarDistritosDistintos conta quantos distritos diferentes existem de fato
// na tabela ruas (em vez de um número fixo no código).
func (r *ruaRepository) ContarDistritosDistintos(ctx context.Context) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&models.Rua{}).
		Distinct("distrito").
		Where("distrito IS NOT NULL AND distrito <> ''").
		Count(&total).Error
	return total, err
}
