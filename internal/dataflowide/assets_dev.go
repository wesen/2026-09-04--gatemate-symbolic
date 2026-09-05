//go:build !embed

package dataflowide

import "os"

func asset(name string) ([]byte, error) { return os.ReadFile("internal/dataflowide/assets/" + name) }
