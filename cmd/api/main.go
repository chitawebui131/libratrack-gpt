package main

import (
	_ "libratrack/docs"

	handler "libratrack/internal/handler"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Libratrack API
// @version 1.0
// @description REST API for a library catalog.
// @BasePath /api/v1
func main() {
	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	// CRUD routes for books
	router.GET("/api/v1/books", handler.GetBooks)
	router.POST("/api/v1/books", handler.CreateBook)
	router.GET("/api/v1/books/:id", handler.GetBook)
	router.PUT("/api/v1/books/:id", handler.UpdateBook)
	router.DELETE("/api/v1/books/:id", handler.DeleteBook)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.Run(":8080")
}
