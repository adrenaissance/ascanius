package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/adrenaissance/ascanius"
)

type Server struct {
	Host string
	Port int
}

func main() {
	var err error
	var serverConfig Server

	builder := ascanius.New().
		SetSource("./files/config.json", 100).
		Load(&serverConfig)

	if builder.HasErrs() {
		fmt.Println(builder.Errs())
		return
	}

	addr := fmt.Sprintf("%s:%d", serverConfig.Host, serverConfig.Port)
	fmt.Printf("Starting server on %s...\n", addr)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, world!")
	})

	err = http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
