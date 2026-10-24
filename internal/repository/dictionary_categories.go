package repository

import (
	"gorm.io/gorm"

	"libratrack/internal/model"
)

func (r *gormDictionaryRepository) ListCategories() ([]model.Category, error) {
	// 1. var items []model.Category. err := r.db.Order("name ASC, id ASC").Find(&items).Error.
	// 2. If err != nil, return nil, translateError(err). Otherwise return items, nil.
	var items []model.Category
	err := r.db.Order("name ASC, id ASC").Find(&items).Error
	if err != nil {
		return nil, translateError(err)
	}
	return items, nil
}

func (r *gormDictionaryRepository) GetCategory(id uint) (*model.Category, error) {
	// 1. var c model.Category. err := r.db.First(&c, id).Error.
	// 2. If err != nil, return nil, translateError(err). Otherwise return &c, nil.
	var c model.Category
	err := r.db.First(&c, id).Error
	if err != nil {
		return nil, translateError(err)
	}
	return &c, nil
}

func (r *gormDictionaryRepository) CreateCategory(c *model.Category) error {
	// err := r.db.Create(c).Error. If err != nil return translateError(err). Otherwise nil.
	err := r.db.Create(c).Error
	if err != nil {
		return translateError(err)
	}
	return nil
}

func (r *gormDictionaryRepository) UpdateCategory(c *model.Category) error {
	// 1. res := r.db.Model(&model.Category{}).Where("id = ?", c.ID).Select("Name").Updates(c).
	// 2. If res.Error != nil, return translateError(res.Error).
	// 3. If res.RowsAffected == 0, return ErrNotFound.
	// 4. Return nil.
	res := r.db.Model(&model.Category{}).Where("id = ?", c.ID).Select("Name").Updates(c)
	if res.Error != nil {
		return translateError(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteCategory removes a category. A category used by a book is protected by the foreign key (ErrInvalidReference).
func (r *gormDictionaryRepository) DeleteCategory(id uint) error {
	// 1. res := r.db.Delete(&model.Category{}, id). If res.Error != nil return translateError(res.Error).
	res := r.db.Delete(&model.Category{}, id)
	if res.Error != nil {
		return translateError(res.Error)
	}
	// 2. If res.RowsAffected == 0 return ErrNotFound. Otherwise return nil.
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func NewDictionaryRepository(db *gorm.DB) DictionaryRepository {
	// Return &gormDictionaryRepository{db: db}.
	return &gormDictionaryRepository{db: db}
}
