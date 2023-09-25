package main

import (
	"fmt"
	"net/http"
)

type serverConfig struct {
	Port int
}

func newServer(config serverConfig) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", ping)
	mux.HandleFunc("/books", books)
	mux.HandleFunc("/books/", bookById)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: mux,
	}
	return &server
}
