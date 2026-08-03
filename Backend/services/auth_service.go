package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/empresa/rotas-entrega/models"
	"github.com/empresa/rotas-entrega/repositories"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	maxTentativasFalhas = 5
	duracaoBloqueio     = 15 * time.Minute
)

var (
	ErrCredenciaisInvalidas = errors.New("matrícula ou senha inválida")
	ErrContaBloqueada       = errors.New("conta temporariamente bloqueada por excesso de tentativas, tente novamente mais tarde")
	ErrSenhaAtualInvalida   = errors.New("senha atual incorreta")
	ErrMatriculaEmUso       = errors.New("matrícula já possui usuário cadastrado")
)

type LoginDTO struct {
	Matricula string `json:"matricula"`
	Senha     string `json:"senha"`
}

type LoginResponse struct {
	Token           string `json:"token"`
	Papel           string `json:"papel"`
	Matricula       string `json:"matricula"`
	SenhaProvisoria bool   `json:"senha_provisoria"`
}

type NovoUsuarioDTO struct {
	Matricula       string `json:"matricula"`
	SenhaTemporaria string `json:"senha_temporaria"`
	Papel           string `json:"papel"` // "admin" ou "colaborador"
	ColaboradorID   *uint  `json:"colaborador_id"`
}

type TrocarSenhaDTO struct {
	SenhaAtual string `json:"senha_atual"`
	SenhaNova  string `json:"senha_nova"`
}

// Claims é o payload do JWT: identifica quem está logado e com que papel,
// pra middleware e handlers (ex: gravar em HistoricoAlteracao.Usuario) usarem sem consultar o banco de novo.
type Claims struct {
	UsuarioID uint   `json:"usuario_id"`
	Matricula string `json:"matricula"`
	Papel     string `json:"papel"`
	jwt.RegisteredClaims
}

type ResetarSenhaDTO struct {
	SenhaTemporaria string `json:"senha_temporaria"`
}

type UsuarioResumo struct {
	ID              uint       `json:"id"`
	Matricula       string     `json:"matricula"`
	Papel           string     `json:"papel"`
	SenhaProvisoria bool       `json:"senha_provisoria"`
	Bloqueado       bool       `json:"bloqueado"`
	CriadoEm        time.Time  `json:"criado_em"`
}

type AuthService interface {
	Login(ctx context.Context, dto LoginDTO, ip string) (*LoginResponse, error)
	TrocarSenha(ctx context.Context, usuarioID uint, dto TrocarSenhaDTO) error
	CriarUsuario(ctx context.Context, dto NovoUsuarioDTO) (*models.Usuario, error)
	ListUsuarios(ctx context.Context) ([]UsuarioResumo, error)
	ResetarSenha(ctx context.Context, usuarioID uint, dto ResetarSenhaDTO) error
}

type authService struct {
	usuarioRepo repositories.UsuarioRepository
	acessoRepo  repositories.AcessoLoginRepository
	jwtSecret   string
	jwtHoras    int
}

func NewAuthService(usuarioRepo repositories.UsuarioRepository, acessoRepo repositories.AcessoLoginRepository, jwtSecret string, jwtHoras int) AuthService {
	return &authService{usuarioRepo: usuarioRepo, acessoRepo: acessoRepo, jwtSecret: jwtSecret, jwtHoras: jwtHoras}
}

func (s *authService) Login(ctx context.Context, dto LoginDTO, ip string) (*LoginResponse, error) {
	matricula := strings.TrimSpace(dto.Matricula)

	usuario, err := s.usuarioRepo.FindByMatricula(ctx, matricula)
	if err != nil {
		// Mesmo com matrícula inexistente, registra a tentativa — ajuda a
		// detectar varredura/força bruta de matrículas.
		s.registrarAcesso(ctx, matricula, false, ip)
		return nil, ErrCredenciaisInvalidas
	}

	if usuario.BloqueadoAte != nil && usuario.BloqueadoAte.After(time.Now()) {
		s.registrarAcesso(ctx, matricula, false, ip)
		return nil, ErrContaBloqueada
	}

	if err := bcrypt.CompareHashAndPassword([]byte(usuario.SenhaHash), []byte(dto.Senha)); err != nil {
		usuario.TentativasFalhas++
		if usuario.TentativasFalhas >= maxTentativasFalhas {
			bloqueio := time.Now().Add(duracaoBloqueio)
			usuario.BloqueadoAte = &bloqueio
			usuario.TentativasFalhas = 0
		}
		_ = s.usuarioRepo.Update(ctx, usuario)
		s.registrarAcesso(ctx, matricula, false, ip)
		return nil, ErrCredenciaisInvalidas
	}

	// Login certo: zera tentativas e bloqueio, se houver.
	usuario.TentativasFalhas = 0
	usuario.BloqueadoAte = nil
	if err := s.usuarioRepo.Update(ctx, usuario); err != nil {
		return nil, err
	}
	s.registrarAcesso(ctx, matricula, true, ip)

	token, err := s.gerarToken(usuario)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:           token,
		Papel:           usuario.Papel,
		Matricula:       usuario.Matricula,
		SenhaProvisoria: usuario.SenhaProvisoria,
	}, nil
}

func (s *authService) TrocarSenha(ctx context.Context, usuarioID uint, dto TrocarSenhaDTO) error {
	usuario, err := s.usuarioRepo.FindByID(ctx, usuarioID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(usuario.SenhaHash), []byte(dto.SenhaAtual)); err != nil {
		return ErrSenhaAtualInvalida
	}

	if len(dto.SenhaNova) < 8 {
		return errors.New("a nova senha precisa ter pelo menos 8 caracteres")
	}
	if dto.SenhaNova == usuario.Matricula {
		return errors.New("a nova senha não pode ser igual à matrícula")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(dto.SenhaNova), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	usuario.SenhaHash = string(hash)
	usuario.SenhaProvisoria = false
	return s.usuarioRepo.Update(ctx, usuario)
}

// CriarUsuario só deve ser chamado por rota protegida a admin (feito no handler/middleware).
func (s *authService) CriarUsuario(ctx context.Context, dto NovoUsuarioDTO) (*models.Usuario, error) {
	matricula := strings.TrimSpace(dto.Matricula)
	if matricula == "" || dto.SenhaTemporaria == "" {
		return nil, errors.New("matrícula e senha temporária são obrigatórias")
	}

	if _, err := s.usuarioRepo.FindByMatricula(ctx, matricula); err == nil {
		return nil, ErrMatriculaEmUso
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	papel := strings.TrimSpace(dto.Papel)
	if papel != models.PapelAdmin {
		papel = models.PapelColaborador
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(dto.SenhaTemporaria), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	usuario := &models.Usuario{
		Matricula:       matricula,
		SenhaHash:       string(hash),
		Papel:           papel,
		ColaboradorID:   dto.ColaboradorID,
		SenhaProvisoria: true,
	}

	if err := s.usuarioRepo.Create(ctx, usuario); err != nil {
		return nil, err
	}
	return usuario, nil
}

func (s *authService) ListUsuarios(ctx context.Context) ([]UsuarioResumo, error) {
	usuarios, err := s.usuarioRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	resumos := make([]UsuarioResumo, 0, len(usuarios))
	for _, u := range usuarios {
		resumos = append(resumos, UsuarioResumo{
			ID:              u.ID,
			Matricula:       u.Matricula,
			Papel:           u.Papel,
			SenhaProvisoria: u.SenhaProvisoria,
			Bloqueado:       u.BloqueadoAte != nil && u.BloqueadoAte.After(time.Now()),
			CriadoEm:        u.CreatedAt,
		})
	}
	return resumos, nil
}

// ResetarSenha é a versão "admin reseta a senha de outra pessoa" —
// diferente de TrocarSenha, não pede a senha atual (o admin não sabe e não
// deveria saber). Só quem já passou por ExigirAdmin no handler chega aqui.
func (s *authService) ResetarSenha(ctx context.Context, usuarioID uint, dto ResetarSenhaDTO) error {
	if len(dto.SenhaTemporaria) < 8 {
		return errors.New("a senha temporária precisa ter pelo menos 8 caracteres")
	}

	usuario, err := s.usuarioRepo.FindByID(ctx, usuarioID)
	if err != nil {
		return errors.New("usuário não encontrado")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(dto.SenhaTemporaria), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	usuario.SenhaHash = string(hash)
	usuario.SenhaProvisoria = true
	usuario.TentativasFalhas = 0
	usuario.BloqueadoAte = nil
	return s.usuarioRepo.Update(ctx, usuario)
}

func (s *authService) gerarToken(usuario *models.Usuario) (string, error) {
	claims := Claims{
		UsuarioID: usuario.ID,
		Matricula: usuario.Matricula,
		Papel:     usuario.Papel,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.jwtHoras) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *authService) registrarAcesso(ctx context.Context, matricula string, sucesso bool, ip string) {
	_ = s.acessoRepo.Create(ctx, &models.AcessoLogin{
		Matricula: matricula,
		Sucesso:   sucesso,
		IP:        ip,
	})
}
