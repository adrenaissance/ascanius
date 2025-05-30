package ascanius

import (
	"encoding/json"
	"os"
)

type JsonSource struct {
	name     string
	path     string
	priority int
}

func NewJsonSource(path string, name string, priority int) *JsonSource {
	if name == "" {
		name = path
	}
	return &JsonSource{
		name:     name,
		path:     path,
		priority: priority,
	}
}

func (j *JsonSource) Load() (map[string]any, error) {
	result := make(map[string]any)

	bytes, err := os.ReadFile(j.path)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (j *JsonSource) Name() string {
	return j.name
}

func (j *JsonSource) SetName(name string) {
	j.name = name
}

func (j *JsonSource) Priority() int {
	return j.priority
}

func (j *JsonSource) SetPriority(p int) {
	j.priority = p
}

func (j *JsonSource) Type() string {
	return "json"
}
