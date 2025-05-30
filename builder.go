package ascanius

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type Builder struct {
	sources   []Source
	mapSource map[string]map[string]any
	errs      []error
	envPrefix string
	envSep    string
}

func New() *Builder {
	return &Builder{
		sources:   []Source{},
		mapSource: make(map[string]map[string]any),
		envPrefix: "APP",
		envSep:    "__",
	}
}

func (b *Builder) SetEnvPrefix(prefix string) *Builder {
	b.envPrefix = prefix
	return b
}

func (b *Builder) SetEnvSeparator(sep string) *Builder {
	b.envSep = sep
	return b
}

func (b *Builder) SetSource(name string, priority int) *Builder {
	nameLower := strings.ToLower(name)

	switch {
	case nameLower == "env":
		b.sources = append(b.sources, NewEnvSource("env", priority, WithPrefix(b.envPrefix), WithSeparator(b.envSep)))

	case strings.HasPrefix(filepath.Base(nameLower), ".env"):
		b.sources = append(b.sources, NewEnvSource(nameLower, priority, WithPrefix(b.envPrefix), WithSeparator(b.envSep)))

	case strings.HasSuffix(filepath.Base(nameLower), ".json"):
		b.sources = append(b.sources, NewJsonSource(nameLower, "", priority))

	case strings.HasSuffix(filepath.Base(nameLower), ".toml"):
		b.sources = append(b.sources, NewTomlSource(nameLower, "", priority))

	case strings.HasSuffix(filepath.Base(nameLower), ".yaml"):
		b.sources = append(b.sources, NewYamlSource(nameLower, "", priority))

	case !strings.Contains(name, "."):
		b.errs = append(b.errs, fmt.Errorf("no source type provided for %s", name))

	default:
		b.errs = append(b.errs, fmt.Errorf("unsupported source type for %s", name))
	}

	return b
}

func (b *Builder) LoadSection(target any, section string) *Builder {
	var err error
	if target == nil {
		b.errs = append(b.errs, errors.New("target cannot be nil"))
		return b
	}

	sort.SliceStable(b.sources, func(i, j int) bool {
		return b.sources[i].Priority() < b.sources[j].Priority()
	})

	merged := make(map[string]any)

	for _, src := range b.sources {
		name := src.Name()

		var data map[string]any
		if cached, ok := b.mapSource[name]; ok {
			data = cached
		} else {
			data, err = src.Load()
			if err != nil {
				b.errs = append(b.errs, err)
				continue
			}
			data = normalizeKeysToSnakeCase(data)
			b.mapSource[name] = data
		}

		merged = mergeMaps(merged, data)
	}

	sectionKey := toSnakeCase(section)
	if sectionData, ok := merged[sectionKey]; ok {
		if sectionMap, ok := sectionData.(map[string]any); ok {
			err := b.applyValues(target, sectionMap)
			if err != nil {
				b.errs = append(b.errs, err)
			}
			return b
		}
	}
	b.errs = append(b.errs, fmt.Errorf("section '%s' not found", sectionKey))
	return b
}

func (b *Builder) Load(target any) *Builder {
	var err error
	if target == nil {
		b.errs = append(b.errs, errors.New("target cannot be nil"))
		return b
	}

	sort.SliceStable(b.sources, func(i, j int) bool {
		return b.sources[i].Priority() < b.sources[j].Priority()
	})

	merged := make(map[string]any)

	for _, src := range b.sources {
		name := src.Name()

		var data map[string]any
		if cached, ok := b.mapSource[name]; ok {
			data = cached
		} else {
			data, err = src.Load()
			if err != nil {
				b.errs = append(b.errs, err)
				continue
			}
			data = normalizeKeysToSnakeCase(data)
			b.mapSource[name] = data
		}

		merged = mergeMaps(merged, data)
	}

	err = b.applyValues(target, merged)
	if err != nil {
		b.errs = append(b.errs, err)
	}
	return b
}

func (b *Builder) applyValues(target any, data map[string]any) error {
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return errors.New("target must be a non-nil pointer to a struct")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return errors.New("target must point to a struct")
	}

	typ := val.Type()
	structNameKey := toSnakeCase(typ.Name())
	if sectionData, ok := data[structNameKey]; ok {
		if sectionMap, ok := sectionData.(map[string]any); ok {
			return b.applyValues(target, sectionMap)
		}
	}

	for i := range typ.NumField() {
		field := typ.Field(i)
		fieldVal := val.Field(i)
		if !fieldVal.CanSet() {
			continue
		}

		cfgTag := field.Tag.Get("cfg")
		if cfgTag == "" {
			cfgTag = toSnakeCase(field.Name)
		}

		value, exists := data[cfgTag]
		if !exists {
			if defVal := field.Tag.Get("def"); defVal != "" {
				if parsedVal, err := parseDefault(defVal, fieldVal.Type()); err == nil {
					fieldVal.Set(parsedVal)
				}
			}
			continue
		}

		if fieldVal.Kind() == reflect.Struct {
			if subMap, ok := value.(map[string]any); ok {
				if err := b.applyValues(fieldVal.Addr().Interface(), subMap); err != nil {
					return fmt.Errorf("error in section %s: %w", cfgTag, err)
				}
				continue
			}
		}

		if parsed, err := convertValue(value, fieldVal.Type()); err == nil {
			fieldVal.Set(parsed)
		}
	}
	return nil
}

func convertValue(value any, targetType reflect.Type) (reflect.Value, error) {
	val := reflect.ValueOf(value)

	if val.Type().AssignableTo(targetType) {
		return val, nil
	}

	if val.Type().ConvertibleTo(targetType) {
		return val.Convert(targetType), nil
	}

	b, err := json.Marshal(value)
	if err != nil {
		return reflect.Value{}, err
	}

	ptr := reflect.New(targetType)
	if err := json.Unmarshal(b, ptr.Interface()); err != nil {
		return reflect.Value{}, err
	}

	return ptr.Elem(), nil
}

func parseDefault(def string, t reflect.Type) (reflect.Value, error) {
	switch t.Kind() {
	case reflect.String:
		return reflect.ValueOf(def), nil

	case reflect.Bool:
		v, err := strconv.ParseBool(def)
		return reflect.ValueOf(v), err

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(def, 10, 64)
		return reflect.ValueOf(v).Convert(t), err

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(def, 10, 64)
		return reflect.ValueOf(v).Convert(t), err

	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(def, 64)
		return reflect.ValueOf(v).Convert(t), err

	case reflect.Slice:
		if t.Elem().Kind() == reflect.String {
			return reflect.ValueOf(strings.Split(def, ",")), nil
		}
		return reflect.Value{}, fmt.Errorf("unsupported slice type")

	default:
		return reflect.Value{}, fmt.Errorf("unsupported kind: %s", t.Kind())
	}
}

func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) && (unicode.IsLower(rune(s[i-1])) || (i+1 < len(s) && unicode.IsLower(rune(s[i+1])))) {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

func (b *Builder) HasErrs() bool {
	return len(b.errs) > 0 && b.errs != nil
}

func (b *Builder) Errs() []error {
	return b.errs
}

func (b *Builder) Panic() {
	if b.HasErrs() {
		panic(b.Errs)
	}
}

func normalizeKeysToSnakeCase(data any) map[string]any {
	switch m := data.(type) {
	case map[string]any:
		out := make(map[string]any)
		for k, v := range m {
			newKey := toSnakeCase(k)
			switch nested := v.(type) {
			case map[string]any:
				out[newKey] = normalizeKeysToSnakeCase(nested)
			case []any:
				out[newKey] = normalizeArray(nested)
			default:
				out[newKey] = v
			}
		}
		return out
	default:
		return map[string]any{}
	}
}

func normalizeArray(arr []any) []any {
	for i, item := range arr {
		switch v := item.(type) {
		case map[string]any:
			arr[i] = normalizeKeysToSnakeCase(v)
		case []any:
			arr[i] = normalizeArray(v)
		}
	}
	return arr
}

func mergeMaps(dst, src map[string]any) map[string]any {
	for k, v := range src {
		if existing, ok := dst[k]; ok {
			existingMap, isMap := existing.(map[string]any)
			newMap, isNewMap := v.(map[string]any)
			if isMap && isNewMap {
				dst[k] = mergeMaps(existingMap, newMap)
				continue
			}
		}
		dst[k] = v
	}
	return dst
}
