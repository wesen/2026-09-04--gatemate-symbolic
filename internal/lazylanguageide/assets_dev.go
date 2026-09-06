//go:build !embed

package lazylanguageide

import "os"

func asset(name string) ([]byte, error) {
	return os.ReadFile("internal/lazylanguageide/assets/" + name)
}
