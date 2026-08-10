package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort        string
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	DBSSLMode         string
	JWTSecret         string
	JWTExpiracaoHoras int
	CORSOrigins       string
	// URL do Worker do Cloudflare que faz proxy pro Groq (chave da Groq
	// mora só lá, o Backend nunca vê ela). Opcional — sem isso, o Zé Rota
	// fica desligado, o resto do site funciona normal.
	ZeRotaWorkerURL string
}

func Load() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		ServerPort:        getEnv("SERVER_PORT", "8080"),
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBUser:            getEnv("DB_USER", "postgres"),
		DBPassword:        getEnv("DB_PASSWORD", "postgres"),
		DBName:            getEnv("DB_NAME", "rotas_entrega"),
		DBSSLMode:         getEnv("DB_SSLMODE", "disable"),
		JWTSecret:         getEnv("JWT_SECRET", ""),
		JWTExpiracaoHoras: getEnvInt("JWT_EXPIRACAO_HORAS", 3),
		// Lista separada por vírgula, sem espaço, ex: "https://devsaulo.com.br,http://localhost:5173"
		CORSOrigins:     getEnv("CORS_ORIGINS", "http://localhost:5173"),
		ZeRotaWorkerURL: getEnv("ZE_ROTA_WORKER_URL", ""),
	}

	if cfg.JWTSecret == "" {
		slog.Warn("JWT_SECRET não definido — usando valor de desenvolvimento, NUNCA use isso em produção")
		cfg.JWTSecret = "dev-secret-trocar-em-producao"
	}

	return cfg
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort, c.DBSSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	if defaultValue == "" {
		slog.Warn("variável de ambiente não definida", "key", key)
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	valor := os.Getenv(key)
	if valor == "" {
		return defaultValue
	}
	numero, err := strconv.Atoi(valor)
	if err != nil {
		slog.Warn("valor inválido pra variável numérica, usando padrão", "key", key, "valor", valor)
		return defaultValue
	}
	return numero
}
