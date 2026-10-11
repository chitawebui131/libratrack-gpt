package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetBooks handles GET requests for retrieving all books.
func GetBooks(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetBooks"})
}

// CreateBook handles POST requests for creating a new book.
func CreateBook(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "CreateBook"})
}

// GetBook handles GET requests for retrieving a single book by ID.
func GetBook(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetBook"})
}

// UpdateBook handles PUT requests for updating an existing book by ID.
func UpdateBook(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "UpdateBook"})
}

// DeleteBook handles DELETE requests for removing a book by ID.
func DeleteBook(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "DeleteBook"})
}
