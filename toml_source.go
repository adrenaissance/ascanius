package ascanius

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

type TomlSource struct {
	name     string
	path     string
	priority int
}

func NewTomlSource(path string, name string, priority int) *TomlSource {
	if name == "" {
		name = path
	}
	return &TomlSource{
		name:     name,
		path:     path,
		priority: priority,
	}
}

func (t *TomlSource) Load() map[string]any {
	result := make(map[string]any)

	bytes, err := os.ReadFile(t.path)
	if err != nil {
		return result
	}

	err = toml.Unmarshal(bytes, &result)
	if err != nil {
		return make(map[string]any)
	}

	return result
}

func (t *TomlSource) Name() string        { return t.name }
func (t *TomlSource) SetName(name string) { t.name = name }
func (t *TomlSource) Priority() int       { return t.priority }
func (t *TomlSource) SetPriority(p int)   { t.priority = p }
func (t *TomlSource) Type() string        { return "toml" }
