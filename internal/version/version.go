package version

import (
	_ "embed"
	"strings"
)

//go:embed version.txt
var raw string

var Version = strings.TrimSpace(raw)
