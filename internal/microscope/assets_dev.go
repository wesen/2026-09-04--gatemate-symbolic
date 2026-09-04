//go:build !embed

package microscope

import "os"

func asset(name string) ([]byte, error) { return os.ReadFile("internal/microscope/assets/" + name) }
