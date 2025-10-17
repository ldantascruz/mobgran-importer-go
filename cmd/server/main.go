package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"mobgran-importer-go/internal/config"
	"mobgran-importer-go/internal/handlers"
	"mobgran-importer-go/internal/services"
	"mobgran-importer-go/pkg/database"
	_ "mobgran-importer-go/docs"
)

// @title Mobgran Importer API
// @version 1.0
// @description API para importação de ofertas do Mobgran para Supabase
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8081
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

	// Inicializar cliente PostgreSQL
	dbClient, err := database.NewClient(
		cfg.DBHost, cfg.DBPort, cfg.DBName, 
		cfg.DBUser, cfg.DBPassword, cfg.DBSSLMode, 
		logger,
	)
	if err != nil {
		log.Fatalf("Erro ao inicializar cliente PostgreSQL: %v", err)
	}
	defer dbClient.Close()

	// Inicializar serviço de importação
	importerService := services.NewMobgranImporter(dbClient, logger)

	// Inicializar handlers
	importerHandler := handlers.NewImporterHandler(importerService, logger)

	// Configurar Gin
	if cfg.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middlewares
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// Rotas de saúde
	router.GET("/health", importerHandler.HealthCheck)
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Mobgran Importer API - Go",
			"version": "1.0.0",
			"status":  "running",
		})
	})

	// Rotas da API
	api := router.Group("/api")
	{
		api.POST("/importar", importerHandler.ImportarOferta)
		api.POST("/validar-url", importerHandler.ValidarURL)
		api.POST("/extrair-uuid", importerHandler.ExtrairUUID)
	}

	// Rota do Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Iniciar servidor
	port := cfg.Port
	logger.WithField("port", port).Info("Iniciando servidor")

	if err := router.Run(":" + port); err != nil {
		logger.WithError(err).Fatal("Erro ao iniciar servidor")
	}
}

// corsMiddleware adiciona headers CORS
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}