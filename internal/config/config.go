package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// Config representa a configuração da aplicação
type Config struct {
	// Servidor
	Port string

	// PostgreSQL Database
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string

	// Supabase
	SupabaseURL        string `mapstructure:"SUPABASE_URL"`
	SupabaseKey        string `mapstructure:"SUPABASE_KEY"`
	SupabaseServiceKey string `mapstructure:"SUPABASE_SERVICE_KEY"`

	// Logging
	LogLevel string

	// Mobgran API
	MobgranAPIURL string
}

// LoadConfig carrega a configuração da aplicação
func LoadConfig() (*Config, error) {
	// Carregar arquivo .env se existir
	if err := godotenv.Load(); err != nil {
		logrus.Warn("Arquivo .env não encontrado, usando variáveis de ambiente do sistema")
	}

	config := &Config{
		Port:          getEnvOrDefault("PORT", "8080"),
		DBHost:        getEnvOrDefault("DB_HOST", "localhost"),
		DBPort:        getEnvOrDefault("DB_PORT", "5432"),
		DBName:        getEnvOrDefault("DB_NAME", "mobgran"),
		DBUser:        getEnvOrDefault("DB_USER", "mobgran_user"),
		DBPassword:    getEnvOrDefault("DB_PASSWORD", "mobgran_password"),
		DBSSLMode:     getEnvOrDefault("DB_SSLMODE", "disable"),
		SupabaseURL:        getEnvOrDefault("SUPABASE_URL", ""),
		SupabaseKey:        getEnvOrDefault("SUPABASE_KEY", ""),
		SupabaseServiceKey: getEnvOrDefault("SUPABASE_SERVICE_KEY", ""),
		LogLevel:      getEnvOrDefault("LOG_LEVEL", "info"),
		MobgranAPIURL: getEnvOrDefault("MOBGRAN_API_URL", "https://api.mobgran.com.br/api/v1/ofertas/"),
	}

	// Validar configurações obrigatórias do PostgreSQL
	if config.DBHost == "" {
		return nil, fmt.Errorf("DB_HOST é obrigatório")
	}

	if config.DBName == "" {
		return nil, fmt.Errorf("DB_NAME é obrigatório")
	}

	if config.DBUser == "" {
		return nil, fmt.Errorf("DB_USER é obrigatório")
	}

	if config.DBPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD é obrigatório")
	}

	return config, nil
}

// getEnvOrDefault retorna o valor da variável de ambiente ou um valor padrão
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SetupLogger configura o logger baseado no nível de log
func SetupLogger(logLevel string) *logrus.Logger {
	logger := logrus.New()

	// Configurar formato
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})

	// Configurar nível
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logger.Warn("Nível de log inválido, usando 'info'")
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	return logger
}