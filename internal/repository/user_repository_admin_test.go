package repository

import (
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"libratrack/internal/model"
)

func TestUserUpdateProfile(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user1 := model.User{
		Email:        "user1@example.com",
		PasswordHash: "hash",
		FullName:     "User One",
		Role:         model.RoleReader,
	}
	if err := db.Create(&user1).Error; err != nil {
		t.Fatal(err)
	}

	user2 := model.User{
		Email:        "user2@example.com",
		PasswordHash: "hash",
		FullName:     "User Two",
		Role:         model.RoleReader,
	}
	if err := db.Create(&user2).Error; err != nil {
		t.Fatal(err)
	}

	// UpdateProfile of the first one to a new email and name succeeds
	err := repo.UpdateProfile(user1.ID, "newemail@example.com", "New Name")
	if err != nil {
		t.Fatal(err)
	}

	// FindByID shows the new values
	foundUser, err := repo.FindByID(user1.ID)
	if err != nil {
		t.Fatal(err)
	}
	if foundUser.Email != "newemail@example.com" {
		t.Fatalf("expected email 'newemail@example.com', got '%s'", foundUser.Email)
	}
	if foundUser.FullName != "New Name" {
		t.Fatalf("expected full_name 'New Name', got '%s'", foundUser.FullName)
	}

	// UpdateProfile of the first one to the email of the second gives ErrConflict
	err = repo.UpdateProfile(user1.ID, user2.Email, "New Name")
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}

	// UpdateProfile(999999, ...) gives ErrNotFound
	err = repo.UpdateProfile(999999, "newemail@example.com", "New Name")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUserSetPasswordHashAndDisabled(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := model.User{
		Email:        "user@example.com",
		PasswordHash: "hash",
		FullName:     "User",
		Role:         model.RoleReader,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}

	// SetPasswordHash changes password_hash
	newHash := "newhash"
	err := repo.SetPasswordHash(user.ID, newHash)
	if err != nil {
		t.Fatal(err)
	}

	foundUser, err := repo.FindByID(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if foundUser.PasswordHash != newHash {
		t.Fatalf("expected password_hash '%s', got '%s'", newHash, foundUser.PasswordHash)
	}

	// SetDisabled(id, true) then FindByID: Disabled true
	err = repo.SetDisabled(user.ID, true)
	if err != nil {
		t.Fatal(err)
	}
	foundUser, err = repo.FindByID(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !foundUser.Disabled {
		t.Fatalf("expected disabled true, got false")
	}

	// SetDisabled(id, false) then Disabled false
	err = repo.SetDisabled(user.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	foundUser, err = repo.FindByID(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if foundUser.Disabled {
		t.Fatalf("expected disabled false, got true")
	}

	// SetDisabled(999999, true) gives ErrNotFound
	err = repo.SetDisabled(999999, true)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUserDeleteWithoutLoans(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := model.User{
		Email:        "user@example.com",
		PasswordHash: "hash",
		FullName:     "User",
		Role:         model.RoleReader,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}

	// Delete succeeds
	err := repo.Delete(user.ID)
	if err != nil {
		t.Fatal(err)
	}

	// FindByID returns ErrNotFound
	_, err = repo.FindByID(user.ID)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}

	// Delete(999999) gives ErrNotFound
	err = repo.Delete(999999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUserDeleteWithLoanHistoryIsBlocked(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	category := mustCategory(t, db, "Test Category")
	author := mustAuthor(t, db, "Test Author")
	book := mustBook(t, db, "Test Book", "1234567890", category.ID, []uint{author.ID})

	cp := model.Copy{
		BookID:      book.ID,
		Status:      model.CopyStatusAvailable,
		InventoryNo: "INV00000000001",
	}
	if err := db.Create(&cp).Error; err != nil {
		t.Fatal(err)
	}

	user := model.User{
		Email:        "user@example.com",
		PasswordHash: "hash",
		FullName:     "User",
		Role:         model.RoleReader,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}

	loan := model.Loan{
		CopyID:     cp.ID,
		UserID:     user.ID,
		BorrowedAt: db.NowFunc(),
		DueAt:      db.NowFunc().Add(14 * 24 * time.Hour),
	}
	if err := db.Create(&loan).Error; err != nil {
		t.Fatal(err)
	}

	// Delete returns ErrInvalidReference and the user still exists
	err := repo.Delete(user.ID)
	if !errors.Is(err, ErrInvalidReference) {
		t.Fatalf("expected ErrInvalidReference, got %v", err)
	}

	// User still exists
	_, err = repo.FindByID(user.ID)
	if err != nil {
		t.Fatal(err)
	}
}
