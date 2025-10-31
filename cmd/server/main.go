package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"mobgran-importer-go/internal/config"
	"mobgran-importer-go/internal/handlers"
	"mobgran-importer-go/internal/middleware"
	"mobgran-importer-go/internal/services"
	"mobgran-importer-go/pkg/database"
	_ "mobgran-importer-go/docs"
)

// @title Mobgran Importer API
// @version 1.0
// @description API para importação de ofertas do Mobgran com PostgreSQL - HOT RELOAD FUNCIONANDO! 🔥
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @schemes http https

func main() {
	// Carregar configuração
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Erro ao carregar configuração: %v", err)
	}

	// Configurar logger
	logger := config.SetupLogger(cfg.LogLevel)

	// Inicializar cliente PostgreSQL com migrations automáticas
	connString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBUser, cfg.DBPassword, cfg.DBSSLMode)
	
	dbClient, err := database.NewPostgresClient(connString)
	if err != nil {
		log.Fatalf("Erro ao inicializar cliente PostgreSQL: %v", err)
	}
	defer dbClient.Close()

	// Executar migrations automáticas
	log.Println("🔄 Chamando RunMigrations()...")
	if err := dbClient.RunMigrations(); err != nil {
		log.Fatalf("Erro ao executar migrations: %v", err)
	}
	log.Println("✅ RunMigrations() concluído com sucesso!")

	// Inicializar serviços de autenticação
	authService := services.NewAuthService(dbClient)

	// Inicializar handlers
	authHandler := handlers.NewAuthHandler(authService)

	// Configurar Gin
	if cfg.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middlewares
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.RecoveryMiddleware())

	// Rotas de saúde
	router.GET("/health", func(c *gin.Context) {
		// Verificar saúde do banco de dados
		if err := dbClient.HealthCheck(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "database connection failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "healthy",
			"database": "connected",
			"version":  "1.0.0",
		})
	})

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Mobgran Importer API - PostgreSQL 🔥 HOT RELOAD ATIVO!",
			"version": "1.0.0",
			"status":  "running",
		})
	})

	// Rotas de autenticação
	auth := router.Group("/auth")
	{
		auth.POST("/registrar", authHandler.Registrar)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/logout", middleware.AuthMiddleware(), authHandler.Logout)
		auth.GET("/perfil", middleware.AuthMiddleware(), authHandler.Perfil)
		auth.PUT("/perfil", middleware.AuthMiddleware(), authHandler.AtualizarPerfil)
	}

	// Rota do Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Iniciar servidor
	port := cfg.Port
	logger.WithField("port", port).Info("Iniciando servidor PostgreSQL")

	if err := router.Run(":" + port); err != nil {
		logger.WithError(err).Fatal("Erro ao iniciar servidor")
	}
}