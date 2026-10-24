package repository

import "libratrack/internal/model"

func (r *gormDictionaryRepository) ListAuthors() ([]model.Author, error) {
	var items []model.Author
  err := r.db.Order("full_name ASC, id ASC").Find(&items).Error
	if err != nil {
		return nil, translateError(err)
	}
	return items, nil
}

func (r *gormDictionaryRepository) GetAuthor(id uint) (*model.Author, error) {
	var a model.Author
	err := r.db.First(&a, id).Error
	if err != nil {
		return nil, translateError(err)
	}
	return &a, nil
}

func (r *gormDictionaryRepository) CreateAuthor(a *model.Author) error {
	err := r.db.Create(a).Error
	if err != nil {
		return translateError(err)
	}
	return nil
}

func (r *gormDictionaryRepository) UpdateAuthor(a *model.Author) error {
	res := r.db.Model(&model.Author{}).Where("id = ?", a.ID).Select("FullName", "Bio").Updates(a)
	if res.Error != nil {
		return translateError(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteAuthor removes an author. An author used by a book is protected by the foreign key (ErrInvalidReference).
func (r *gormDictionaryRepository) DeleteAuthor(id uint) error {
	res := r.db.Delete(&model.Author{}, id)
	if res.Error != nil {
		return translateError(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
