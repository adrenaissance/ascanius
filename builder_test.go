package ascanius

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TlsConfig struct {
	Ca         string `def:"/etc/ssl/ca"`
	Key        string `def:"/etc/ssl/server.key"`
	Cert       string `def:"/etc/ssl/server.crt"`
	Disabled   bool   `def:"true"`
	EnableMtls bool   `def:"false"`
}

type ServerConfig struct {
	HttpPort  uint16 `def:"8080"`
	HttpsPort uint16 `def:"8443"`
	Host      string `def:"localhost"`
	Name      string `def:"server"`
	Tls       TlsConfig
}

type MongoConfig struct {
	Scheme         string `def:"mongodb"`
	Host           string `def:"localhost"`
	Port           uint16 `def:"27017"`
	Username       string `def:"username"`
	Password       string `def:"password"`
	Database       string `def:"test"`
	Collection     string `def:"test"`
	Params         string `def:"?ssl=true"`
	ReplicaSet     string `def:""`
	ConnectTimeout uint64 `def:"10"`
	ReadPreference string `def:"primary"`
}

type LogConfig struct {
	Level           string   `def:"error"`
	Outputs         []string `def:"stdout,file:logs/app.log"`
	MessageField    string   `def:"body"`
	TimestampFormat string   `def:"2006-01-02T15:04:05.999999999Z07:00"`
	Resource        string   `def:""`
}

type AppConfig struct {
	Log    LogConfig
	Mongo  MongoConfig
	Server ServerConfig
}

func setTestEnv(t *testing.T) {
	envs := map[string]string{
		"APP__LOG__LEVEL":               "debug",
		"APP__LOG__OUTPUTS":             `["stdout"]`,
		"APP__MONGO__COLLECTION":        "test-collection",
		"APP__MONGO__DATABASE":          "test-db",
		"APP__MONGO__HOST":              "test.mongo.local",
		"APP__MONGO__OPERATION_TIMEOUT": "10s",
		"APP__MONGO__PARAMS":            "?ssl=true",
		"APP__MONGO__PASSWORD":          "dummy-password",
		"APP__MONGO__SCHEME":            "mongodb+srv",
		"APP__MONGO__USERNAME":          "dummy-user",
		"APP__SERVER__HOST":             "127.0.0.1",
		"APP__SERVER__HTTPS_PORT":       "9443",
		"APP__SERVER__HTTP_PORT":        "9000",
		"APP__SERVER__NAME":             "dummy-server",
		"APP__SERVER__TLS__DISABLE_SSL": "true",
		"PLACEHOLDER":                   "dummy-placeholder",
	}

	for k, v := range envs {
		t.Setenv(k, v)
	}
}

func TestBuilderLoadEnv(t *testing.T) {
	setTestEnv(t)

	var cfgData AppConfig
	builder := New().
		EnvPrefix("APP").
		EnvSeparator("__").
		Source("env", 100)

	builder.Load(&cfgData)
	require.Equal(t, builder.HasErrs(), false)

	require.Equal(t, "debug", cfgData.Log.Level)
	require.Len(t, cfgData.Log.Outputs, 1)
	require.Equal(t, "stdout", cfgData.Log.Outputs[0])

	require.Equal(t, "test-db", cfgData.Mongo.Database)
	require.Equal(t, "test-collection", cfgData.Mongo.Collection)
	require.Equal(t, "dummy-user", cfgData.Mongo.Username)
	require.Equal(t, "dummy-password", cfgData.Mongo.Password)
	require.Equal(t, "test.mongo.local", cfgData.Mongo.Host)

	require.Equal(t, "127.0.0.1", cfgData.Server.Host)
	require.Equal(t, uint16(9000), cfgData.Server.HttpPort)
	require.Equal(t, uint16(9443), cfgData.Server.HttpsPort)
	require.Equal(t, "dummy-server", cfgData.Server.Name)
	require.True(t, cfgData.Server.Tls.Disabled)
}

func TestDotEnvLoading(t *testing.T) {
	type Config struct {
		Log struct {
			Level string
		}
		Mongo struct {
			Database string
		}
		Server struct {
			Host string
		}
	}

	var cfg Config

	builder := New().
		EnvPrefix("APP").
		EnvSeparator("__").
		Source("./files/.env", 100).
		Source("./files/.env.development", 1000).
		Load(&cfg)

	require.Equal(t, builder.HasErrs(), false)
	require.Equal(t, "error", cfg.Log.Level)
	require.Equal(t, "test-db", cfg.Mongo.Database)
	require.Equal(t, "127.0.0.1", cfg.Server.Host)
}

func TestTomlSource(t *testing.T) {
	type A struct {
		Mongo MongoConfig `cfg:"mongo"`
	}

	var a A
	builder := New().Source("./files/config.toml", 100).Load(&a)
	require.Equal(t, builder.HasErrs(), false)

	config := a.Mongo

	require.Equal(t, "mongodb+srv", config.Scheme)
	require.Equal(t, "mongo.example.com", config.Host)
	require.Equal(t, uint16(27018), config.Port)
	require.Equal(t, "toml-user", config.Username)
	require.Equal(t, "toml-pass", config.Password)
	require.Equal(t, "toml-db", config.Database)
	require.Equal(t, "toml-collection", config.Collection)
	require.Equal(t, "?retryWrites=true&w=majority", config.Params)
	require.Equal(t, "rs0", config.ReplicaSet)
	require.Equal(t, uint64(20), config.ConnectTimeout)
	require.Equal(t, "secondary", config.ReadPreference)
}

func TestJsonSource(t *testing.T) {
	type A struct {
		Mongo MongoConfig `cfg:"mongo"`
	}

	var a A
	builder := New().Source("./files/mongo.json", 100).Load(&a)
	require.Equal(t, builder.HasErrs(), false)

	config := a.Mongo
	require.Equal(t, "mongodb+srv", config.Scheme)
	require.Equal(t, "mongo.example.com", config.Host)
	require.Equal(t, uint16(27018), config.Port)
	require.Equal(t, "toml-user", config.Username)
	require.Equal(t, "toml-pass", config.Password)
	require.Equal(t, "toml-db", config.Database)
	require.Equal(t, "toml-collection", config.Collection)
	require.Equal(t, "?retryWrites=true&w=majority", config.Params)
	require.Equal(t, "rs0", config.ReplicaSet)
	require.Equal(t, uint64(20), config.ConnectTimeout)
	require.Equal(t, "secondary", config.ReadPreference)
}

func TestYamlSource(t *testing.T) {
	type A struct {
		Mongo MongoConfig `cfg:"mongo"`
	}

	var a A
	builder := New().
		Source("./files/mongo.yaml", 100).
		Load(&a)
	require.Equal(t, builder.HasErrs(), false)
	config := a.Mongo
	require.Equal(t, "mongodb+srv", config.Scheme)
	require.Equal(t, "mongo.example.com", config.Host)
	require.Equal(t, uint16(27018), config.Port)
	require.Equal(t, "toml-user", config.Username)
	require.Equal(t, "toml-pass", config.Password)
	require.Equal(t, "toml-db", config.Database)
	require.Equal(t, "toml-collection", config.Collection)
	require.Equal(t, "?retryWrites=true&w=majority", config.Params)
	require.Equal(t, "rs0", config.ReplicaSet)
	require.Equal(t, uint64(20), config.ConnectTimeout)
	require.Equal(t, "secondary", config.ReadPreference)
}

func TestSingleSectionLoading(t *testing.T) {
	type Mongo struct {
		Scheme         string `def:"mongodb"`
		Host           string `def:"localhost"`
		Port           uint16 `def:"27017"`
		Username       string `def:"username"`
		Password       string `def:"password"`
		Database       string `def:"test"`
		Collection     string `def:"test"`
		Params         string `def:"?ssl=true"`
		ReplicaSet     string `def:""`
		ConnectTimeout uint64 `def:"10"`
		ReadPreference string `def:"primary"`
	}

	var cfg Mongo
	var cfg2 Mongo

	builder := New().
		Source("./files/mongo.yaml", 100)

	builder.Load(&cfg)
	builder.LoadSection(&cfg2, "MONGO")

	require.Equal(t, builder.HasErrs(), false)
	require.Equal(t, builder.HasErrs(), false)

	require.Equal(t, cfg, cfg2)

	config := cfg
	require.Equal(t, "mongodb+srv", config.Scheme)
	require.Equal(t, "mongo.example.com", config.Host)
	require.Equal(t, uint16(27018), config.Port)
	require.Equal(t, "toml-user", config.Username)
	require.Equal(t, "toml-pass", config.Password)
	require.Equal(t, "toml-db", config.Database)
	require.Equal(t, "toml-collection", config.Collection)
	require.Equal(t, "?retryWrites=true&w=majority", config.Params)
	require.Equal(t, "rs0", config.ReplicaSet)
	require.Equal(t, uint64(20), config.ConnectTimeout)
	require.Equal(t, "secondary", config.ReadPreference)
}

func TestBadFiles(t *testing.T) {
	type Section struct {
		Name string `def:"default"`
	}

	var cfg Section
	builder := New().
		Source("./files/bad.toml", 100).
		Source("./files/bad.yaml", 101).
		Source("./files/bad.json", 102).
		Load(&cfg)

	require.True(t, builder.HasErrs())
	require.Equal(t, cfg.Name, "default")
	require.Equal(t, len(builder.Errs()), 3)
}

func TestFlatConfiguration(t *testing.T) {
	type ServerCfg struct {
		Host string
		Port int
		Name string
	}

	var cfg ServerCfg
	builder := New().
		Source("./files/flat.toml", 1).
		Source("./files/flat.json", 2).
		Source("./files/flat.yaml", 3).
		Load(&cfg)

	require.False(t, builder.HasErrs())
	assert.Equal(t, cfg.Host, "localhost")
	assert.Equal(t, cfg.Port, 9090)
	assert.Equal(t, cfg.Name, "hello")
}

func TestMultipleFlatConfigs(t *testing.T) {
	type Mongo struct {
		Scheme         string `def:"mongodb"`
		Host           string `def:"localhost" cfg:"mongo_host"`
		Port           uint16 `def:"27017" cfg:"mongo_port"`
		Username       string `def:"username"`
		Password       string `def:"password"`
		Database       string `def:"test"`
		Collection     string `def:"test"`
		Params         string `def:"?ssl=true"`
		ReplicaSet     string `def:""`
		ConnectTimeout uint64 `def:"10"`
		ReadPreference string `def:"primary"`
	}

	type ServerCfg struct {
		Host string
		Port int
	}

	var server ServerCfg
	var mongo Mongo
	builder := New().
		Source("./files/multi-flat.toml", 1).
		Load(&server).
		Load(&mongo)

	require.False(t, builder.HasErrs())
	assert.Equal(t, server.Host, "localhost")
	assert.Equal(t, server.Port, 8080)

	require.Equal(t, "mongodb+srv", mongo.Scheme)
	require.Equal(t, "mongo.example.com", mongo.Host)
	require.Equal(t, uint16(27018), mongo.Port)
	require.Equal(t, "toml-user", mongo.Username)
	require.Equal(t, "toml-pass", mongo.Password)
	require.Equal(t, "toml-db", mongo.Database)
	require.Equal(t, "toml-collection", mongo.Collection)
	require.Equal(t, "?retryWrites=true&w=majority", mongo.Params)
	require.Equal(t, "rs0", mongo.ReplicaSet)
	require.Equal(t, uint64(20), mongo.ConnectTimeout)
	require.Equal(t, "secondary", mongo.ReadPreference)
}
