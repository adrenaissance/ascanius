package ascanius

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const (
	ENV_SOURCE_NAME    = "env"
	DOTENV_SOURCE_NAME = ".env"
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
	if s.name == ENV {
		return ENV_SOURCE_NAME
	} else {
		return DOTENV_SOURCE_NAME
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
		prefix:   DEFAULT_ENV_PREFIX,
		sep:      DEFAULT_ENV_SEPARATOR,
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

func (e *EnvSource) Load() (map[string]any, error) {
	flat := make(map[string]string)
	prefix := e.prefix + e.sep

	if e.name == ENV {
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
			return nil, err
		}
		for k, v := range envMap {
			if strings.HasPrefix(k, prefix) {
				key := strings.TrimPrefix(k, prefix)
				flat[strings.ToLower(key)] = v
			}
		}
	}

	return expandEnv(flat, e.sep), nil
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
