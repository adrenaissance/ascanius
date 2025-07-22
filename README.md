# Ascanius

Ascanius is designed to seamlessly aggregate configuration from multiple sources and populate your Go structs with minimal boilerplate.



## Twelve-Factor Configuration Support

Ascanius is fully aligned with the [Twelve-Factor App](https://12factor.net/config) methodology. It encourages storing configuration in the environment and provides flexible, layered overrides across `.env` files, environment variables, and other formats. This makes it ideal for cloud-native applications that require:

- Environment-specific overrides  
- Externalized configuration  
- Immutable builds and deploy-time config injection  

By centralizing and abstracting configuration loading, Ascanius enables you to build portable, environment-agnostic services that follow modern deployment practices.



## Supported Sources and Priority

Ascanius supports the following configuration sources:

- JSON files (`.json`)
- YAML files (`.yaml`. `.yml`)
- TOML files (`.toml`)
- `.env` files (e.g., `.env`, `.env.development`, `.env.staging`, `.env.config`, etc.)
- OS environment variables

Each source can be assigned a **priority**. Higher-priority sources override values from lower-priority ones. Example usage:

```go
ascanius.New().
    Source("config.toml", 1).
    Source("config.yaml", 2).
    Source("config.json", 3).
    Source(".env", 4).
    Source(".env.development", 5).
    Source(".env.config", 6).
    Source("env", 100) // system env vars
    Load(&cfg)
```



## Load Order and Merge Logic

Ascanius performs the following steps internally:

1. Sorts sources by ascending priority.
2. Loads data from each source, normalizing all keys to `snake_case`.
3. Merges data from each source, with higher-priority values overriding lower-priority ones.
4. Applies the final merged configuration to the provided Go struct.



## Environment Variables and `.env` Files

Ascanius supports both `.env` files and actual environment variables. You can mix and prioritize them however you need.

### `.env` Files

`.env` files are parsed using [`godotenv`](https://github.com/joho/godotenv). Ascanius allows loading any file that starts with `.env`, such as:

- `.env`
- `.env.development`
- `.env.config`
- `.env.staging`

Each can be added as a separate source with its own priority:

```go
ascanius.New().
    Source(".env", 1).
    Source(".env.development", 2).
    Source(".env.config", 3)
```

### Environment Variables

To load from OS-level environment variables:

```go
ascanius.New().
    SetSource("env", 100) // system environment
```

Ascanius supports **key mapping using a prefix and separator**, which lets you represent nested structures:

```go
ascanius.New().
    EnvPrefix("APP").
    EnvSeparator("__").
    Source("env", 100)
```

With that setup, the following environment variables:

```env
APP_DB__HOST=localhost
APP_DB__PORT=5432
APP_CACHE__ENABLED=true
```

Are parsed as:

```json
{
  "db": {
    "host": "localhost",
    "port": 5432
  },
  "cache": {
    "enabled": true
  }
}
```



## Field Resolution and Default Values

When applying values to struct fields, Ascanius follows this resolution order:

1. Uses the `cfg` tag if provided, otherwise defaults to the snake_case version of the field name.
2. Looks for a matching key in the merged configuration.
3. Uses the `def` tag as a fallback default.

### With `cfg` Tags

```go
type Config struct {
  DBHost     string `cfg:"db.host"`
  DBPort     int    `cfg:"db.port"`
  CacheOn    bool   `cfg:"cache.enabled" def:"false"`
}
```

### Without `cfg` Tags

```go
type Config struct {
  DbHost   string
  DbPort   int
  Cache    struct {
    Enabled bool
  }
}
```

Both approaches work, and can be mixed depending on your needs.



## Nested Structs and Sections

Ascanius supports deep merging into nested structs. It automatically recurses into sub-structs as needed. You can also load just a portion of your configuration with `LoadSection`:

```go
ascanius.New().
    Source("config.json", 1).
    LoadSection(&cfg, "database")
```

This applies only the `"database"` section of the merged config to `cfg`.

