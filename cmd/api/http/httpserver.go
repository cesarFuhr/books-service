package http

import (
	"fmt"
	"net/http"
)

type ServerConfig struct {
	Port int
}

func NewServer(config ServerConfig, h *BookHandler) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", ping)
	mux.HandleFunc("/books", h.books)
	mux.HandleFunc("/books/", h.bookById)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: mux,
	}
	return &server
}

/* Tests the http server connection.  */
func ping(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	if method == http.MethodGet {
		w.WriteHeader(http.StatusNoContent)
		return
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
