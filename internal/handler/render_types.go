package handler

import "libratrack/internal/model"

// PageData is the value every page template receives.
type PageData struct {
	Title     string
	User      *model.User
	CSRFToken string
	Flash     string
	FlashArg  int64
	Data      any
}

// Pager describes pagination links.
type Pager struct {
	Page       int
	TotalPages int
	Total      int64
	PrevURL    string
	NextURL    string
}

// ErrorPage is the data of the error page.
type ErrorPage struct {
	Code int
}
