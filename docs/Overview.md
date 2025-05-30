# 🔄 How Ascanius Loads Configuration Variables

Ascanius is built to seamlessly aggregate configuration from multiple sources and populate your Go structs with minimal boilerplate. Here's how the variable loading mechanism works:

---

## 📥 Supported Sources (with Priority)

Ascanius supports the following configuration sources:

- JSON files (`.json`)
- YAML files (`.yaml`)
- TOML files (`.toml`)
- `.env` files (e.g., `.env`, `.env.development`)
- OS environment variables

Each source can be given a **priority**, and higher-priority sources override lower ones. For example:

```go
ascanius.New().
    SetSource("config.toml", 1).
    SetSource("config.yaml", 2).
    SetSource("config.json", 3).
    SetSource(".env.development", 4).
    SetSource("env", 100).
    Load(&cfg)
```

In this example:
- Values from `env` (priority 100) will override values from `.env.development` (priority 4)
- `.env.development` will override values from `config.json` (priority 3), and so on.

---

## 🔄 Load Order

Sources are **sorted by ascending priority** (lowest number first). Internally, Ascanius:

1. Sorts the sources by priority.
2. Loads data from each source and normalizes all keys to `snake_case`.
3. Merges the sources into a single map, from lowest to highest priority.
4. Applies the resulting merged configuration to your target struct.

---

## 📦 Environment Variables and `.env` Files

Environment-based configs are handled in two ways:

- `SetSource("env", ...)` reads actual environment variables.
- `SetSource(".env.development", ...)` reads key-value pairs from `.env` files using `godotenv`.

Both are parsed using a prefix (`APP`) and separator (`__`) to build nested config maps. For example:

```env
APP_DB__HOST=localhost
APP_DB__PORT=5432
```

These become:

```json
{
  "db": {
    "host": "localhost",
    "port": 5432
  }
}
```

You can customize the prefix and separator:

```go
ascanius.New().
    SetEnvPrefix("MYAPP").
    SetEnvSeparator("_")
```

---

## 🧠 Variable Resolution & Defaults

When loading into structs, Ascanius tries to resolve each field by:

1. Looking for a field's `cfg` tag, or using the snake_case version of its name.
2. Matching the field to the merged configuration map.
3. Using the `def` tag as a fallback default if the value wasn't set.

Example:

```go
type Config struct {
    Name        string `cfg:"app_name"`
    Description string `def:"some default description"`
}
```

---

## 🧩 Nested Structs & Sections

Ascanius also supports loading into nested structs. It recurses into struct fields, loading nested configuration maps as needed. Additionally, you can load a specific **section** by name:

```go
ascanius.New().
    SetSource("config.json", 1).
    LoadSection(&cfg, "database")
```

This will try to match a `"database"` key in the merged config and apply it to `cfg`.

---

## ✅ Summary

- ✅ Multiple sources with custom priority  
- ✅ Snake-case normalization and deep merging  
- ✅ Support for nested structs, environment variables, `.env` files  
- ✅ `cfg` tags for custom keys, `def` for defaults  

Ascanius gives you simple yet powerful configuration loading with just a few lines of code.
