package handler

import (
	"fmt"
	"html/template"
	"time"
)

// templateFuncs returns the functions available in every template.
func templateFuncs(loc *time.Location) template.FuncMap {
	return template.FuncMap{
		"date": func(t time.Time) string {
			// Return t.In(loc).Format("2006-01-02").
			return t.In(loc).Format("2006-01-02")
		},
		"datetime": func(t time.Time) string {
			// Return t.In(loc).Format("2006-01-02 15:04").
			return t.In(loc).Format("2006-01-02 15:04")
		},
		"datePtr": func(t *time.Time) string {
			// If t == nil return "". Otherwise return t.In(loc).Format("2006-01-02").
			if t == nil {
				return ""
			}
			return t.In(loc).Format("2006-01-02")
		},
		"add": func(a, b int) int {
			// Return a + b.
			return a + b
		},
		"sub": func(a, b int) int {
			// Return a - b.
			return a - b
		},
		// statusClass takes any value because the status may be a named string type.
		"statusClass": func(v any) string {
			// 1. s := fmt.Sprint(v).
			// 2. Return "available" for "available", "borrowed" for "borrowed", "lost" for "lost", "repair" for "under_repair", and "" for anything else.
			s := fmt.Sprint(v)
			switch s {
			case "available":
				return "available"
			case "borrowed":
				return "borrowed"
			case "lost":
				return "lost"
			case "under_repair":
				return "repair"
			default:
				return ""
			}
		},
	}
}
