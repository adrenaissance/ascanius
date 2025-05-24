# Environment Variables and `.env` File Loading

This document explains how environment variables and `.env` files are loaded and processed,  
including how to customize prefixes and separators for nested configuration.

---

## How Environment Variables Are Loaded

By default, the loader reads all OS environment variables that start with a specified **prefix** (default: `APP`) followed by a **separator** (default: `__`).
For example, if the prefix is `APP` and the separator is `__`, the loader considers environment variables like:

```
APP__DATABASE__HOST=localhost
APP__DATABASE__PORT=5432
APP__FEATURE_ENABLED=true
```

**Only environment variables matching this pattern are included.**

### Out of the box support for json string in env vars
Lots of the times we also have environment variables that have json values. This library would handle that case.


### Example: Mapping Environment Variables to Go Struct

Given the environment variables:

```
APP__DATABASE__HOST=localhost
APP__DATABASE__PORT=5432
APP__FEATURE_FLAGS__ENABLE_LOGGING=true
APP__SERVER__TIMEOUT=30
```

Define a Go struct to receive these values:

```go
type Config struct {
    Database struct {
        Host string `cfg:"host"`
        Port int    `cfg:"port"`
    } `cfg:"database"`
    FeatureFlags struct {
        EnableLogging bool `cfg:"enable_logging"`
    } `cfg:"feature_flags"`
    Server struct {
        Timeout int `cfg:"timeout"`
    } `cfg:"server"`
}
```

Load the environment variables into the struct with:

```go
var cfg Config
builder := ascanius.New().
    SetEnvPrefix("APP").
    SetEnvSeparator("__").
    SetSource("env", 10) // 10 is the priority

err := builder.Load(&cfg)
if err != nil {
    panic(err)
}
```

After loading, `cfg` will contain:

```go
Config{
    Database: struct {
        Host string
        Port int
    }{
        Host: "localhost",
        Port: 5432,
    },
    FeatureFlags: struct {
        EnableLogging bool
    }{
        EnableLogging: true,
    },
    Server: struct {
        Timeout int
    }{
        Timeout: 30,
    },
}
```

---

## Loading `.env` Files Automatically

- If a source is specified with a `.env` filename (e.g., `.env` or `.env.local`), the loader reads that file using a dotenv parser.
- The `.env` file can contain key-value pairs just like environment variables:

```
APP__DATABASE__HOST=db.local
APP__DATABASE__PORT=5432
APP__FEATURE_ENABLED=false
```

- These values are loaded and merged with any environment variables, respecting source priority.

---

## Prefix and Separator Explained

### Prefix (`APP` by default)

- The **prefix** filters which environment variables or `.env` entries to include.
- Only keys starting with the prefix plus the separator are processed.
- For example, with prefix `MYAPP`, the loader processes:

```
MYAPP__DATABASE__HOST=localhost
```

**Variables without the prefix are ignored.**

### Separator (`__` by default)

- The **separator** defines how keys are split into nested structures.
- For example, the variable:

```
APP__DATABASE__HOST=localhost
```

becomes a nested map:

```json
{
  "database": {
    "host": "localhost"
  }
}
```

### Nested keys with different separator example

If you customize the separator to `%%`, environment variables like:

```
MYAPP%%SERVER%%PORT=8080
MYAPP%%FEATURE_FLAGS%%ENABLED=true
```

become:

```json
{
  "server": {
    "port": 8080
  },
  "feature_flags": {
    "enabled": true
  }
}
```

---

## Customize Prefix and Separator

You can customize prefix and separator in your configuration loading code:

```go
builder.SetEnvPrefix("MYAPP").SetEnvSeparator("%%")
```

Ensure your environment variables or `.env` file keys match these settings.

---

## Summary

| Step                     | Description                                          |
|--------------------------|------------------------------------------------------|
| Filter by prefix         | Only keys starting with `${PREFIX}${SEPARATOR}` are included |
| Strip prefix             | The prefix and separator are removed from keys       |
| Split by separator       | Remaining key parts split into nested levels         |
| Parse values             | Values parsed from strings into native types if possible  |
| Merge sources            | `.env` files and OS environment variables merged by priority |

---
