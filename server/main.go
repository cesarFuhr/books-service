package main

import (
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)

}

func main() {
	http.HandleFunc("/ping", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
