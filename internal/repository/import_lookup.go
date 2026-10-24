package repository

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"libratrack/internal/model"
)

// findOrCreateCategory returns the id of the category with this name (case-insensitive). It creates it when missing.
func findOrCreateCategory(tx *gorm.DB, name string) (uint, error) {
	// 1. var c model.Category. err := tx.Where("lower(name) = lower(?)", name).First(&c).Error.
	var c model.Category
	err := tx.Where("lower(name) = lower(?)", name).First(&c).Error
	// 2. If err == nil, return c.ID, nil. If !errors.Is(err, gorm.ErrRecordNotFound), return 0, err.
	if err == nil {
		return c.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	// 3. c = model.Category{Name: name}. If err := tx.Create(&c).Error; err != nil, return 0, err.
	c = model.Category{Name: name}
	if err := tx.Create(&c).Error; err != nil {
		return 0, err
	}
	// 4. Return c.ID, nil.
	return c.ID, nil
}

// findOrCreateAuthors returns one author per name (case-insensitive match), creating the missing ones.
func findOrCreateAuthors(tx *gorm.DB, names []string) ([]model.Author, error) {
	// 1. out := []model.Author{}; seen := map[string]bool{}.
	out := []model.Author{}
	seen := map[string]bool{}
	// 2. For every name: key := strings.ToLower(name); if seen[key] continue; seen[key] = true.
	//      var a model.Author. err := tx.Where("lower(full_name) = lower(?)", name).First(&a).Error.
	//      If errors.Is(err, gorm.ErrRecordNotFound): a = model.Author{FullName: name}; if err := tx.Create(&a).Error; err != nil return nil, err.
	//      Else if err != nil return nil, err.
	//      Append a to out.
	for _, name := range names {
		key := strings.ToLower(name)
		if seen[key] {
			continue
		}
		seen[key] = true

		var a model.Author
		err := tx.Where("lower(full_name) = lower(?)", name).First(&a).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			a = model.Author{FullName: name}
			if err := tx.Create(&a).Error; err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	// 3. Return out, nil.
	return out, nil
}
