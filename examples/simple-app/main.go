package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/adrenaissance/ascanius"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ServerConfig struct {
	Host string
	Port uint16
}

type LogConfig struct {
	Level string
}

type AppConfig struct {
	Server ServerConfig
	Log    LogConfig
}

func main() {
	var appcfg AppConfig

	b := ascanius.
		New().
		SetSource("./configs/config.toml", 100).
		Load(&appcfg)

	if b.HasErrs() {
		fmt.Println(b.Errs())
	}

	level, err := zerolog.ParseLevel(strings.ToLower(appcfg.Log.Level))
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	log.Info().Msgf("Log level set to %s", level.String())

	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		log.Info().
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Msg("Received HTTP request")

		w.Write([]byte("hello world!"))
	})

	addr := fmt.Sprintf("%s:%d", appcfg.Server.Host, appcfg.Server.Port)
	log.Info().Msgf("Starting server at %s", addr)

	err = http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}
