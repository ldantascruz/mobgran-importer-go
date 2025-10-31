package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// PostgresClient é o cliente para PostgreSQL
type PostgresClient struct {
	DB *sql.DB
}

// NewPostgresClient cria uma nova conexão com PostgreSQL
func NewPostgresClient(connString string) (*PostgresClient, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir conexão: %w", err)
	}

	// Configura pool de conexões
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Testa a conexão
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("erro ao conectar ao PostgreSQL: %w", err)
	}

	log.Println("✅ Conectado ao PostgreSQL com sucesso!")

	return &PostgresClient{DB: db}, nil
}

// Close fecha a conexão com o banco
func (c *PostgresClient) Close() error {
	return c.DB.Close()
}

// RunMigrations executa todas as migrations pendentes
func (c *PostgresClient) RunMigrations() error {
	log.Println("🚀 INICIANDO EXECUÇÃO DE MIGRATIONS - MÉTODO CHAMADO!")
	log.Println("🔄 Iniciando execução de migrations...")

	// Cria tabela de controle de migrations
	if err := c.createMigrationsTable(); err != nil {
		return fmt.Errorf("erro ao criar tabela de migrations: %w", err)
	}
	log.Println("✅ Tabela de migrations criada/verificada")

	// Lista arquivos de migration
	log.Println("🔍 Tentando ler diretório de migrations...")
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		log.Printf("❌ Erro ao ler diretório de migrations: %v", err)
		return fmt.Errorf("erro ao ler diretório de migrations: %w", err)
	}
	log.Printf("📁 Encontrados %d arquivos de migration", len(entries))

	// Ordena por nome para garantir ordem de execução
	var filenames []string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".sql") {
			filenames = append(filenames, entry.Name())
			log.Printf("📄 Migration encontrada: %s", entry.Name())
		}
	}
	sort.Strings(filenames)

	// Executa cada migration
	for _, filename := range filenames {
		log.Printf("🔄 Processando migration: %s", filename)
		if err := c.runMigration(filename); err != nil {
			return fmt.Errorf("erro ao executar migration %s: %w", filename, err)
		}
	}

	log.Println("✅ Migrations executadas com sucesso!")
	return nil
}

// createMigrationsTable cria a tabela de controle de migrations
func (c *PostgresClient) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) UNIQUE NOT NULL,
			executed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := c.DB.Exec(query)
	return err
}

// runMigration executa uma migration específica se ainda não foi executada
func (c *PostgresClient) runMigration(filename string) error {
	log.Printf("🔍 Verificando se migration %s já foi executada...", filename)

	// Verifica se já foi executada
	var exists bool
	err := c.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE filename = $1)",
		filename,
	).Scan(&exists)
	if err != nil {
		log.Printf("❌ Erro ao verificar migration %s: %v", filename, err)
		return err
	}

	if exists {
		log.Printf("⏭️  Migration %s já executada, pulando...", filename)
		return nil
	}

	log.Printf("🚀 Executando migration %s...", filename)

	// Lê o arquivo SQL
	content, err := migrationsFS.ReadFile("migrations/" + filename)
	if err != nil {
		log.Printf("❌ Erro ao ler arquivo %s: %v", filename, err)
		return err
	}
	log.Printf("📖 Conteúdo da migration %s lido com sucesso (%d bytes)", filename, len(content))

	// Executa em uma transação
	tx, err := c.DB.Begin()
	if err != nil {
		log.Printf("❌ Erro ao iniciar transação para %s: %v", filename, err)
		return err
	}
	defer tx.Rollback()

	// Executa o SQL
	log.Printf("▶️  Executando migration: %s", filename)
	if _, err := tx.Exec(string(content)); err != nil {
		log.Printf("❌ Erro ao executar SQL da migration %s: %v", filename, err)
		return fmt.Errorf("erro ao executar SQL: %w", err)
	}

	// Registra na tabela de controle
	log.Printf("✅ Migration %s registrada como executada", filename)
	if _, err := tx.Exec(
		"INSERT INTO schema_migrations (filename) VALUES ($1)",
		filename,
	); err != nil {
		log.Printf("❌ Erro ao registrar migration %s: %v", filename, err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("❌ Erro ao confirmar transação da migration %s: %v", filename, err)
		return err
	}

	log.Printf("✅ Migration %s executada com sucesso!", filename)
	return nil
}

// Query executa uma query SELECT
func (c *PostgresClient) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return c.DB.Query(query, args...)
}

// QueryRow executa uma query que retorna uma única linha
func (c *PostgresClient) QueryRow(query string, args ...interface{}) *sql.Row {
	return c.DB.QueryRow(query, args...)
}

// Exec executa uma query (INSERT, UPDATE, DELETE)
func (c *PostgresClient) Exec(query string, args ...interface{}) (sql.Result, error) {
	return c.DB.Exec(query, args...)
}

// Transaction executa uma função dentro de uma transação
func (c *PostgresClient) Transaction(fn func(*sql.Tx) error) error {
	tx, err := c.DB.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

// HealthCheck verifica se o banco está saudável
func (c *PostgresClient) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return c.DB.PingContext(ctx)
}
