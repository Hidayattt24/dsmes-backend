package templates

import "embed"

// FS holds all embedded HTML email templates.
//
//go:embed *.html
var FS embed.FS
