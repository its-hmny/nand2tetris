package jack

import (
	_ "embed"
	"encoding/json"
)

//go:embed stdlib.json
var content string

var StandardLibraryABI = map[string]map[string]Subroutine{}

func init() { json.Unmarshal([]byte(content), &StandardLibraryABI) }
