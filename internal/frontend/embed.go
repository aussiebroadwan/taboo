// Package frontend provides embedded frontend assets.
package frontend

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var frontendFS embed.FS

// GetFS returns the frontend filesystem, rooted at dist/.
// Returns an error if the dist directory doesn't exist (frontend not built).
func GetFS() (fs.FS, error) {
	return fs.Sub(frontendFS, "dist")
}
