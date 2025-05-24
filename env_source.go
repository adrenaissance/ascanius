package ascanius

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func (s EnvSource) Name() string {
	return s.name
}
func (s EnvSource) Priority() int {
	return s.priority
}

func (s *EnvSource) SetPriority(p int) {
	s.priority = p
}
func (s *EnvSource) SetName(n string) {
	s.name = n
}
func (s EnvSource) Type() string {
	if s.name == "env" {
		return "env"
	} else {
		return "env_file"
	}
}

type EnvSource struct {
	name     string // "env" means OS env vars; anything else is treated as a .env file path
	priority int
	prefix   string
	sep      string
}

func NewEnvSource(name string, priority int, opts ...func(*EnvSource)) *EnvSource {
	e := &EnvSource{
		name:     name,
		priority: priority,
		prefix:   "APP",
		sep:      "__",
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func WithPrefix(prefix string) func(*EnvSource) {
	return func(e *EnvSource) {
		e.prefix = prefix
	}
}

func WithSeparator(sep string) func(*EnvSource) {
	return func(e *EnvSource) {
		e.sep = sep
	}
}

func (e *EnvSource) Load() map[string]any {
	flat := make(map[string]string)
	prefix := e.prefix + e.sep

	if e.name == "env" {
		for _, env := range os.Environ() {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) != 2 {
				continue
			}
			k, v := parts[0], parts[1]
			if strings.HasPrefix(k, prefix) {
				key := strings.TrimPrefix(k, prefix)
				flat[strings.ToLower(key)] = v
			}
		}
	} else {
		envMap, err := godotenv.Read(e.name)
		if err != nil {
			fmt.Printf("[ascanius][%s]: %v\n", e.name, err)
			return nil
		}
		for k, v := range envMap {
			if strings.HasPrefix(k, prefix) {
				key := strings.TrimPrefix(k, prefix)
				flat[strings.ToLower(key)] = v
			}
		}
	}

	return expandEnv(flat, e.sep)
}

func expandEnv(flat map[string]string, sep string) map[string]any {
	root := map[string]any{}

	for key, raw := range flat {
		parts := strings.Split(key, sep)
		current := root

		for i, part := range parts {
			if i == len(parts)-1 {
				var val any
				if err := json.Unmarshal([]byte(raw), &val); err == nil {
					current[part] = val
				} else {
					current[part] = raw
				}
			} else {
				if _, ok := current[part]; !ok {
					current[part] = map[string]any{}
				}
				current = current[part].(map[string]any)
			}
		}
	}

	return root
}
