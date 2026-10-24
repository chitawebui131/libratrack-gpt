package handler

import (
	"net/http"
	"net/url"
	"strconv"
)

// BuildPager computes the page count and the previous/next links.
// The links copy the current query string and replace only "page".
func BuildPager(r *http.Request, page, limit int, total int64) Pager {
	// 1. If page < 1, set page = 1. If limit < 1, set limit = 1.
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 1
	}

	// 2. totalPages := int((total + int64(limit) - 1) / int64(limit)). If totalPages < 1, set totalPages = 1.
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if totalPages < 1 {
		totalPages = 1
	}

	// 3. link := func(p int) string { q := url.Values{}; for k, v := range r.URL.Query() { q[k] = v }; q.Set("page", strconv.Itoa(p)); return r.URL.Path + "?" + q.Encode() }.
	link := func(p int) string {
		q := url.Values{}
		for k, v := range r.URL.Query() {
			q[k] = v
		}
		q.Set("page", strconv.Itoa(p))
		return r.URL.Path + "?" + q.Encode()
	}

	// 4. p := Pager{Page: page, TotalPages: totalPages, Total: total}.
	p := Pager{Page: page, TotalPages: totalPages, Total: total}

	// 5. If page > 1, set p.PrevURL = link(page - 1). If page < totalPages, set p.NextURL = link(page + 1).
	if page > 1 {
		p.PrevURL = link(page - 1)
	}
	if page < totalPages {
		p.NextURL = link(page + 1)
	}

	// 6. Return p.
	return p
}
