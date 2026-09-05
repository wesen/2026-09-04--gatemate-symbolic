//go:build !embed

package lazyide

import "os"

func asset(name string) ([]byte, error) { return os.ReadFile("internal/lazyide/assets/" + name) }
