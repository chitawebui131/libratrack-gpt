package repository

import (
	"libratrack/internal/model"
)

// UpdateProfile changes the email and the full name. A duplicate email gives ErrConflict.
func (r *gormUserRepository) UpdateProfile(id uint, email, fullName string) error {
	res := r.db.Model(&model.User{}).Where("id = ?", id).Updates(map[string]interface{}{"email": email, "full_name": fullName})
	if res.Error != nil {
		return translateError(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// SetPasswordHash stores a new bcrypt hash.
func (r *gormUserRepository) SetPasswordHash(id uint, hash string) error {
	res := r.db.Model(&model.User{}).Where("id = ?", id).Updates(map[string]interface{}{"password_hash": hash})
	if res.Error != nil {
		return translateError(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// SetDisabled switches the account on or off.
func (r *gormUserRepository) SetDisabled(id uint, disabled bool) error {
	res := r.db.Model(&model.User{}).Where("id = ?", id).Updates(map[string]interface{}{"disabled": disabled})
	if res.Error != nil {
		return translateError(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes a user. A user with loan history gives ErrInvalidReference (the foreign key of loans is RESTRICT).
func (r *gormUserRepository) Delete(id uint) error {
	res := r.db.Delete(&model.User{}, id)
	if res.Error != nil {
		return translateError(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
