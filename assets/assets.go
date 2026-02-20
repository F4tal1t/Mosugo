// Package assets provides embedded application resources.
package assets

import "embed"

//go:embed *
var FS embed.FS
