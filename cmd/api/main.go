package main

import (
	_ "libratrack/docs"
	"log"
	"os"

	"libratrack/internal/handler"

	"libratrack/internal/repository"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Libratrack API
// @version 1.0
// @description REST API for a library catalog.
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.
func main() {
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {

		dsn = "postgres://libratrack:libratrack@localhost:5432/libratrack?sslmode=disable"
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "migrations"
	}
	if err := repository.RunMigrations(dsn, migrationsPath); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	// 1. db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{}). On error log.Fatalf and return.
	db, err := repository.NewPostgresDB(dsn)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// 2. err = db.AutoMigrate(&model.Book{}). On error log.Fatalf and return.
	// err = repository.AutoMigrate(db)
	// if err != nil {
	// 	log.Fatalf("failed to migrate database: %v", err)
	// }

	// 3. sec, err := buildSecurity(db). On error log.Fatalf and return.
	sec, err := buildSecurity(db)
	if err != nil {
		log.Fatalf("security setup failed: %v", err)
	}

	bookRepo := repository.NewBookRepository(db)
	bookHandler := handler.NewBookHandler(bookRepo)

	authorRepo := repository.NewAuthorRepository(db)
	authorHandler := handler.NewAuthorHandler(authorRepo)

	categoryRepo := repository.NewCategoryRepository(db)
	categoryHandler := handler.NewCategoryHandler(categoryRepo)

	copyRepo := repository.NewCopyRepository(db)
	copyHandler := handler.NewCopyHandler(copyRepo)
	copyAvailabilityHandler := handler.NewCopyAvailabilityHandler(copyRepo)

	api := router.Group("/api/v1")
	// 4. authed, staff, admin := registerSecurityRoutes(api, sec). On error log.Fatalf and return.
	authed, staff, admin := registerSecurityRoutes(api, sec)

	// Register on authed (any logged-in user):
	authed.GET("/books/:id/copies", copyHandler.ListByBook)

	// Register on api (public):
	api.GET("/books", bookHandler.List)
	api.GET("/books/:id", bookHandler.GetByID)
	api.GET("/books/:id/availability", copyAvailabilityHandler.Availability)
	api.GET("/authors", authorHandler.List)
	api.GET("/authors/:id", authorHandler.GetByID)
	api.GET("/categories", categoryHandler.List)
	api.GET("/categories/:id", categoryHandler.GetByID)

	// Register on staff (librarian and admin):
	staff.POST("/books", bookHandler.Create)
	staff.PUT("/books/:id", bookHandler.Update)
	staff.POST("/books/:id/copies", copyHandler.Create)
	staff.POST("/authors", authorHandler.Create)
	staff.POST("/categories", categoryHandler.Create)

	// Register on admin (admin only):
	admin.DELETE("/books/:id", bookHandler.Delete)

	// 5. web, err := buildWeb(db, sec). On error log.Fatalf and return.
	web, err := buildWeb(db, sec)
	if err != nil {
		log.Fatalf("web setup failed: %v", err)
	}
	// 6. if err := registerWebRoutes(router, web); err != nil { log.Fatalf("web routes failed: %v", err) }
	if err := registerWebRoutes(router, web); err != nil {
		log.Fatalf("web routes failed: %v", err)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Run(":8080")
}
