package repository

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"libratrack/internal/model"
	"libratrack/testhelper"
)

func TestCopyCreateAndList(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	// Setup: create category, author and book
	category := model.Category{Name: "Fiction"}
	db.Create(&category)
	author := model.Author{Name: "John Doe"}
	db.Create(&author)
	book := model.Book{
		Title:       "Test Book",
		CategoryID:  category.ID,
		AuthorID:    author.ID,
		Description: "Test description",
	}
	db.Create(&book)

	// Create copies
	copyRepo := NewCopyRepository(db)
	copy1, err := copyRepo.Create(model.Copy{
		BookID:      book.ID,
		InventoryNo: "INV001",
		Status:      "",
		Location:    "Shelf A",
		Acquisition: "Purchase",
		Condition:   "Good",
		Notes:       "Test note 1",
	})
	assert.NoError(t, err)
	assert.Equal(t, model.CopyStatusAvailable, copy1.Status)

	copy2, err := copyRepo.Create(model.Copy{
		BookID:      book.ID,
		InventoryNo: "INV002",
		Status:      model.CopyStatusLost,
		Location:    "Shelf B",
	})
	assert.NoError(t, err)
	assert.Equal(t, model.CopyStatusLost, copy2.Status)

	// List copies
	copies, err := copyRepo.ListByBook(book.ID)
	assert.NoError(t, err)
	assert.Len(t, copies, 2)
	assert.Contains(t, copies, *copy1)
	assert.Contains(t, copies, *copy2)
}

func TestCopyCreateUnknownBook(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	copyRepo := NewCopyRepository(db)
	_, err := copyRepo.Create(model.Copy{
		BookID:      9999,
		InventoryNo: "INV001",
	})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestCopyDuplicateInventoryNo(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	// Setup: create category, author and book
	category := model.Category{Name: "Fiction"}
	db.Create(&category)
	author := model.Author{Name: "John Doe"}
	db.Create(&author)
	book := model.Book{
		Title:       "Test Book",
		CategoryID:  category.ID,
		AuthorID:    author.ID,
		Description: "Test description",
	}
	db.Create(&book)

	// Create first copy
	copyRepo := NewCopyRepository(db)
	_, err := copyRepo.Create(model.Copy{
		BookID:      book.ID,
		InventoryNo: "INV001",
	})
	assert.NoError(t, err)

	// Try to create second copy with same inventory number
	_, err = copyRepo.Create(model.Copy{
		BookID:      book.ID,
		InventoryNo: "INV001",
	})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrDuplicatedKey))
}

func TestAvailability(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	// Setup: create category, author and book
	category := model.Category{Name: "Fiction"}
	db.Create(&category)
	author := model.Author{Name: "John Doe"}
	db.Create(&author)
	book := model.Book{
		Title:       "Test Book",
		CategoryID:  category.ID,
		AuthorID:    author.ID,
		Description: "Test description",
	}
	db.Create(&book)

	// Create copies with different statuses
	copyRepo := NewCopyRepository(db)
	copy1, _ := copyRepo.Create(model.Copy{BookID: book.ID, InventoryNo: "INV001", Status: model.CopyStatusAvailable})
	copy2, _ := copyRepo.Create(model.Copy{BookID: book.ID, InventoryNo: "INV002", Status: model.CopyStatusAvailable})
	copy3, _ := copyRepo.Create(model.Copy{BookID: book.ID, InventoryNo: "INV003", Status: model.CopyStatusBorrowed})
	copy4, _ := copyRepo.Create(model.Copy{BookID: book.ID, InventoryNo: "INV004", Status: model.CopyStatusLost})

	// Test availability
	availability, err := copyRepo.Availability(book.ID)
	assert.NoError(t, err)
	assert.Equal(t, 4, availability.TotalCopies)
	assert.Equal(t, 2, availability.AvailableCopies)
}

func TestAvailabilityNoCopies(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	// Setup: create category, author and book
	category := model.Category{Name: "Fiction"}
	db.Create(&category)
	author := model.Author{Name: "John Doe"}
	db.Create(&author)
	book := model.Book{
		Title:       "Test Book",
		CategoryID:  category.ID,
		AuthorID:    author.ID,
		Description: "Test description",
	}
	db.Create(&book)

	copyRepo := NewCopyRepository(db)
	availability, err := copyRepo.Availability(book.ID)
	assert.NoError(t, err)
	assert.Equal(t, 0, availability.TotalCopies)
	assert.Equal(t, 0, availability.AvailableCopies)
}

func TestAvailabilityUnknownBook(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	copyRepo := NewCopyRepository(db)
	_, err := copyRepo.Availability(9999)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}
