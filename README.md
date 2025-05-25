# Ascanius

**Ascanius** is a flexible and reflection-based configuration loader for Go. 
It provides a unified way to load, merge, and apply configurations from multiple sources, all while respecting priority and defaults. 
Inspired by real-world needs, it aims to be minimal yet powerful.

---

## Features

-  Reflection-based automatic mapping
-  Priority-based loading
-  Non-destructive deep merging of configurations
-  Support for struct field defaults using struct tags
-  Multi-source support:
    - Environment variables
    - `.env` files
    - `TOML`
    - `JSON`
    - `YAML`

---

##  Description

Ascanius allows you to define your configuration structs in Go and declaratively map them to values coming from various sources. It provides:

- A fluent `Builder` API for defining config sources
- Smart normalization of keys (camelCase → snake_case)
- Safe, recursive merging of nested maps
- Tag-based customization with `cfg` and `def` annotations

Example usage:

```go
type Config struct {
    Port int    `def:"8080"`
    Name string `cfg:"name" def:"ascanius"`
}

var cfg Config
err := ascanius.New().
    SetEnvPrefix("APP").
    SetEnvSeparator("__").
    SetSource("env", 100).
    SetSource("config.toml", 200).
    Load(&cfg)
```

---

## How It Works

Internally, Ascanius:

1. Parses sources in order of increasing priority (higher value = higher priority).
2. Normalizes all keys to `snake_case` for consistent merging.
3. Merges maps deeply and non-destructively.
4. Applies values to your struct using reflection, including nested structs.
5. Fills in defaults using the `def` tag if no source value is found.

---

## Struct Tags

You can control field behavior using tags:

| Tag     | Purpose                                  |
|----------|-------------------------------------------|
| `cfg`    | Explicit key name in config (optional)    |
| `def`    | Default value if no config source matches |

---

## Supported Sources

| Source       | Description                    |
|--------------|--------------------------------|
| `"env"`      | Load from current environment  |
| `.env` files | Load using dotenv-like format  |
| `.json`      | Load from JSON files           |
| `.toml`      | Load from TOML files           |
| `.yaml`      | Load from YAML files           |

---

## Example

```go
type TlsConfig struct {
    Cert     string `def:"/etc/ssl/cert"`
    Key      string `def:"/etc/ssl/key"`
    Disabled bool   `def:"false"`
}

type ServerConfig struct {
    Port int       `def:"8080"`
    Tls  TlsConfig
}

type AppConfig struct {
    Server ServerConfig
}

var cfg AppConfig
ascanius.New().
    SetEnvPrefix("APP").
    SetEnvSeparator("__").
    SetSource("env", 100).
    SetSource("config.toml", 200).
    Load(&cfg)
```

---

## Tests

Run tests:

```bash
go test ./...
```

---

## License

MIT – use freely, contribute openly.
