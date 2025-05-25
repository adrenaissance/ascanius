# Configuration Loading Overview

This document explains the internal behavior of the configuration loading system,  
focusing on how environment variables and other sources map onto Go structs,  
the role of the `cfg` struct tag, key normalization, and default value handling.

---

## 1. Struct Field Mapping and the `cfg` Tag

- The loader populates configuration values into your Go struct using **reflection**.
- Each exported struct field is matched to a key in the loaded configuration data.
- By default, the loader looks for a `cfg` struct tag on each field to determine the key name.
  
  ```go
  type Config struct {
      DatabaseHost string `cfg:"database_host"`
      Port         int    `cfg:"port"`
  }
  ```
  
- If the `cfg` tag is **missing**, the loader automatically converts the Go field name from **CamelCase to snake_case**.  
  For example:
  
  ```go
  type Config struct {
      DatabaseHost string // defaults to "database_host"
      MaxRetries   int    // defaults to "max_retries"
  }
  ```

---

## 2. Key Normalization

- Input keys from environment variables, `.env` files, JSON, YAML, or TOML sources are **normalized to snake_case** before matching.
- Nested keys are represented as nested maps.
For example, the environment variable `APP__DATABASE__HOST=localhost` is normalized and parsed into the nested structure:
```json
{
  "database": {
    "host": "localhost"
  }
}
```
- This normalization ensures consistent matching between input keys and struct fields, regardless of input source formatting.
## 3. Reflection-Heavy Library

- This configuration loader relies heavily on Go's `reflect` package to:
- Traverse the target struct fields.
- Match and set field values dynamically.
- Recursively apply nested maps to nested structs.
- As a result, the target passed to `Load` **must be a pointer to a non-nil struct**.
- The loader can set fields of basic types, structs, and nested structs as long as they are exported and settable.

---

## 4. Supported Types and Default Values

- Supported basic types include:
- `string`
- `bool`
- Signed and unsigned integers (`int`, `int8`, `int16`, `int32`, `int64`, `uint`, etc.)
- Floating point numbers (`float32`, `float64`)
- Slices of strings (only slices of strings supported for defaults)
- Default values can be specified using the `def` struct tag. Example:

```go
type Config struct {
    Host        string   `cfg:"host" def:"localhost"`
    Port        int      `cfg:"port" def:"8080"`
    EnableCache bool     `cfg:"enable_cache" def:"true"`
    Tags        []string `cfg:"tags" def:"dev,staging"`
}
```

- The loader parses default values from strings, converting them to the proper type.
- For slices of strings, defaults should be comma-separated.

---

## 5. What Happens If a Key Is Missing?

- If a key is missing in all loaded sources and **no default** is provided, the struct field remains zero-valued (e.g., empty string, 0, false).
- If a default is provided via the `def` tag, it is parsed and assigned.
- Nested structs are initialized recursively only if keys or defaults exist for their fields.

---

## 6. Summary

| Concept             | Behavior                                                                                   |
|---------------------|--------------------------------------------------------------------------------------------|
| `cfg` tag           | Specifies the exact key name to map a struct field                                         |
| Missing `cfg` tag   | Field name converted from CamelCase to snake_case for key matching                         |
| Key normalization   | All input keys normalized to snake_case to support uniform matching                        |
| Reflection usage    | Uses reflection to dynamically set exported fields; target must be a non-nil pointer struct |
| Default values      | Provided via `def` tag; supports strings, booleans, ints, floats, and string slices        |
| Unsupported types   | Complex slices (non-string), maps, or custom types require manual unmarshalling or are unsupported for defaults |

---

## Example: Struct with `cfg` and `def` tags

```go
type Config struct {
  DatabaseHost string   `cfg:"database_host" def:"localhost"`
  Port         int      `cfg:"port" def:"5432"`
  EnableCache  bool     `cfg:"enable_cache" def:"true"`
  Tags         []string `cfg:"tags" def:"dev,staging"`
}
```

## Example: Struct without `cfg` tags (automatic snake_case keys)

```go
type Config struct {
  DatabaseHost string
  MaxRetries   int
  DebugMode    bool
}
```

This struct expects keys:
```
database_host
max_retries
debug_mode
```

in the loaded configuration sources.

## Example: Loading Configuration

```go
var cfg Config
builder := ascanius.New()
builder.SetEnvPrefix("APP").SetEnvSeparator("__")
builder.SetSource("env", 10)
err := builder.Load(&cfg)
if err != nil {
    panic(err)
}
```
