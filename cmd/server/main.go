package main

import (
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
// @description API para importação de ofertas do Mobgran para Supabase - HOT RELOAD FUNCIONANDO! 🔥
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

	// Inicializar serviços de autenticação
	authService := services.NewAuthService(dbClient.GetDB())
	produtosService := services.NewProdutosService(dbClient.GetDB())

	// Inicializar handlers
	importerHandler := handlers.NewImporterHandler(importerService, logger)
	authHandler := handlers.NewAuthHandler(authService)
	produtosHandler := handlers.NewProdutosHandler(produtosService)

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
	router.GET("/health", importerHandler.HealthCheck)
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Mobgran Importer API - Go 🔥 HOT RELOAD ATIVO!",
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

	// Rotas de produtos (protegidas por autenticação)
	produtos := router.Group("/produtos")
	produtos.Use(middleware.AuthMiddleware())
	{
		produtos.GET("/cavaletes", produtosHandler.ListarCavaletesDisponiveis)
		produtos.POST("/aprovar", produtosHandler.AprovarProduto)
		produtos.GET("/aprovados", produtosHandler.ListarProdutosAprovados)
		produtos.PUT("/:id", produtosHandler.AtualizarProduto)
		produtos.GET("/:id", produtosHandler.BuscarProduto)
		produtos.DELETE("/:id", produtosHandler.RemoverProduto)
		produtos.GET("/estatisticas", produtosHandler.ObterEstatisticas)
	}

	// Rotas públicas de vitrine
	vitrine := router.Group("/vitrine")
	{
		vitrine.GET("/publica", produtosHandler.ListarVitrinePublica)
	}

	// Rotas administrativas (protegidas por autenticação)
	admin := router.Group("/admin")
	admin.Use(middleware.AuthMiddleware())
	{
		admin.DELETE("/limpar-dados", produtosHandler.LimparTodosRegistros)
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