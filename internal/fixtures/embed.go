package fixtures

import "embed"

//go:embed dev/*.yaml
var devFiles embed.FS
