package web

import "embed"

// FS holds the HTML templates and the static files of the web UI.
//
//go:embed all:templates all:static
var FS embed.FS
