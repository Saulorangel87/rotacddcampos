// Comando standalone pra criar (ou resetar a senha de) um usuário de acesso
// com o papel escolhido — "admin" ou "colaborador". Complementa o
// cmd/seed-admin, que existe só pra resolver o bootstrap do primeiro admin.
//
// Uso:
//
//	go run cmd/criar-usuario/main.go -matricula 54321 -senha "SenhaTemporaria123" -papel colaborador
//	go run cmd/criar-usuario/main.go -matricula 54321 -senha "SenhaTemporaria123" -papel admin
//
// Se a matrícula já existir, atualiza o papel e reseta a senha. A senha
// sempre nasce como provisória — força troca no próximo login.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/empresa/rotas-entrega/config"
	"github.com/empresa/rotas-entrega/database"
	"github.com/empresa/rotas-entrega/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	matricula := flag.String("matricula", "", "matrícula do usuário (obrigatório)")
	senha := flag.String("senha", "", "senha temporária, mínimo 8 caracteres (obrigatório)")
	papel := flag.String("papel", "colaborador", "papel: \"colaborador\" ou \"admin\"")
	flag.Parse()

	if *matricula == "" || len(*senha) < 8 {
		fmt.Println(`uso: go run cmd/criar-usuario/main.go -matricula 54321 -senha "SenhaTemporaria123" -papel colaborador`)
		fmt.Println("a senha precisa ter pelo menos 8 caracteres")
		os.Exit(1)
	}
	if *papel != models.PapelAdmin && *papel != models.PapelColaborador {
		fmt.Printf("papel inválido: %q — use \"admin\" ou \"colaborador\"\n", *papel)
		os.Exit(1)
	}

	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		slog.Error("falha ao conectar no banco", "error", err)
		os.Exit(1)
	}

	if err := database.RunMigrations(db); err != nil {
		slog.Error("falha ao rodar migrations", "error", err)
		os.Exit(1)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(*senha), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("falha ao gerar hash", "error", err)
		os.Exit(1)
	}

	var usuario models.Usuario
	err = db.Where("matricula = ?", *matricula).First(&usuario).Error

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		usuario = models.Usuario{
			Matricula:       *matricula,
			SenhaHash:       string(hash),
			Papel:           *papel,
			SenhaProvisoria: true,
		}
		if err := db.Create(&usuario).Error; err != nil {
			slog.Error("falha ao criar usuário", "error", err)
			os.Exit(1)
		}
		fmt.Printf("Usuário criado: matrícula %s, papel %s. Senha provisória — vai pedir troca no primeiro login.\n", *matricula, *papel)

	case err == nil:
		usuario.SenhaHash = string(hash)
		usuario.Papel = *papel
		usuario.SenhaProvisoria = true
		usuario.TentativasFalhas = 0
		usuario.BloqueadoAte = nil
		if err := db.Save(&usuario).Error; err != nil {
			slog.Error("falha ao atualizar usuário", "error", err)
			os.Exit(1)
		}
		fmt.Printf("Usuário %s atualizado pra papel %s, com senha resetada. Vai pedir troca no próximo login.\n", *matricula, *papel)

	default:
		slog.Error("erro ao consultar usuário", "error", err)
		os.Exit(1)
	}
}
