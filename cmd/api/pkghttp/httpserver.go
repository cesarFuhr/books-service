package pkghttp

import (
	"fmt"
	"net/http"
)

type ServerConfig struct {
	Port int
}

func NewServer(config ServerConfig) *http.Server {
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
