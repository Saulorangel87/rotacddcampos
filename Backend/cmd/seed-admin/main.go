// Comando standalone pra criar (ou resetar a senha de) um usuário admin
// direto no banco, sem precisar da API. Resolve o problema do "primeiro admin":
// o endpoint POST /auth/usuarios só pode ser chamado por quem já é admin.
//
// Uso:
//
//	go run cmd/seed-admin/main.go -matricula 12345 -senha "SenhaTemporaria123"
//
// Se a matrícula já existir, apenas reseta a senha (útil também pra "esqueci
// minha senha" do admin, já que não existe fluxo self-service por design).
// A senha é sempre criada/reposta como provisória — força troca no próximo login.
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
	matricula := flag.String("matricula", "", "matrícula do admin (obrigatório)")
	senha := flag.String("senha", "", "senha temporária, mínimo 8 caracteres (obrigatório)")
	flag.Parse()

	if *matricula == "" || len(*senha) < 8 {
		fmt.Println("uso: go run cmd/seed-admin/main.go -matricula 12345 -senha \"SenhaTemporaria123\"")
		fmt.Println("a senha precisa ter pelo menos 8 caracteres")
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
			Papel:           models.PapelAdmin,
			SenhaProvisoria: true,
		}
		if err := db.Create(&usuario).Error; err != nil {
			slog.Error("falha ao criar admin", "error", err)
			os.Exit(1)
		}
		fmt.Printf("Admin criado: matrícula %s. Senha provisória — vai pedir troca no primeiro login.\n", *matricula)

	case err == nil:
		usuario.SenhaHash = string(hash)
		usuario.Papel = models.PapelAdmin
		usuario.SenhaProvisoria = true
		usuario.TentativasFalhas = 0
		usuario.BloqueadoAte = nil
		if err := db.Save(&usuario).Error; err != nil {
			slog.Error("falha ao resetar senha", "error", err)
			os.Exit(1)
		}
		fmt.Printf("Senha resetada pra matrícula %s (já era admin). Senha provisória — vai pedir troca no próximo login.\n", *matricula)

	default:
		slog.Error("erro ao consultar usuário", "error", err)
		os.Exit(1)
	}
}
