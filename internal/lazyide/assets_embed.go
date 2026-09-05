//go:build embed

package lazyide

import (
	_ "embed"
	"io/fs"
)

//go:embed assets/index.html
var indexHTML []byte

//go:embed assets/app.js
var appJS []byte

//go:embed assets/app.css
var appCSS []byte

func asset(name string) ([]byte, error) {
	switch name {
	case "index.html":
		return indexHTML, nil
	case "app.js":
		return appJS, nil
	case "app.css":
		return appCSS, nil
	default:
		return nil, fs.ErrNotExist
	}
}
