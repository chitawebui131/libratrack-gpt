package repository

import (
	"sort"
	"sync"
	"time"

	"libratrack/internal/model"
)

// BookRepository defines the operations for managing books.
type BookRepository interface