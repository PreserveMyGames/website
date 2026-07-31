package staticfiles

import "embed"

//go:embed all:static
var FS embed.FS
