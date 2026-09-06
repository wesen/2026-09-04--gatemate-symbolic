package lazylang

import _ "embed"

//go:embed shared.lazy
var shared string

//go:embed closure.lazy
var closure string

//go:embed unused.lazy
var unused string

//go:embed cycle.lazy
var cycle string

//go:embed productive.lazy
var productive string

//go:embed squares.lazy
var squares string

func Sources() map[string]string {
	return map[string]string{"shared": shared, "closure": closure, "unused": unused, "cycle": cycle, "productive": productive, "squares": squares}
}
