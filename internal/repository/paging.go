package repository

// pagingBounds returns the offset and the page size.
// Defaults: page 1, limit 20. Maximum limit: 100.
func pagingBounds(page, limit int) (offset, size int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return (page - 1) * limit, limit
}
