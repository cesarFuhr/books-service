package main

import (
	"fmt"
	"log"
	"net/http"
)

func getPing(w http.ResponseWriter, r *http.Request) {
	metodo := r.Method
	fmt.Printf("metodo: %v\n", metodo)
	if metodo == "GET" {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func getBooks(w http.ResponseWriter, r *http.Request) {
	bookslist := []struct {
		name      string
		price     float32
		inventory int
	}{
		{"Book 1", 30.20, 2},
		{"Book 2", 20.30, 1},
		{"Book 3", 32.20, 5},
	}

	stringbookslist := make([]string, len(bookslist))

	var responsebody string

	for i, v := range bookslist {
		stringbookslist[i] = fmt.Sprintln("name:", v.name, "\nprice:", v.price, "\ninventory:", v.inventory, "\n")
		responsebody = fmt.Sprint(responsebody, stringbookslist[i])
	}

	w.Write([]byte(responsebody))
}

func main() {
	http.HandleFunc("/ping", getPing)
	http.HandleFunc("/books", getBooks)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
