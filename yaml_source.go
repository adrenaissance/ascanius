package ascanius

import (
	"os"

	"gopkg.in/yaml.v3"
)

const YAML_SOURCE_NAME = "yaml"

type YamlSource struct {
	name     string
	path     string
	priority int
}

func NewYamlSource(path string, name string, priority int) *YamlSource {
	if name == "" {
		name = path
	}
	return &YamlSource{
		name:     name,
		path:     path,
		priority: priority,
	}
}

func (t *YamlSource) Load() (map[string]any, error) {
	result := make(map[string]any)

	bytes, err := os.ReadFile(t.path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(bytes, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (t *YamlSource) Name() string        { return t.name }
func (t *YamlSource) SetName(name string) { t.name = name }
func (t *YamlSource) Priority() int       { return t.priority }
func (t *YamlSource) SetPriority(p int)   { t.priority = p }
func (t *YamlSource) Type() string        { return YAML_SOURCE_NAME }
