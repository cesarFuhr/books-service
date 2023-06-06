package main

import (
	"encoding/json"
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

	bookslist := []struct { //Struct created to handle data of the list of books
		Name      string  `json:"name"`
		Price     float32 `json:"price"`
		Inventory int     `json:"inventory"`
	}{
		{"Book 1", 30.20, 2},
		{"Book 2", 20.30, 1},
		{"Book 3", 32.20, 5},
	}

	//Encoding in JSON to send through the Writer:
	responsebody, err := json.Marshal(bookslist)
	if err != nil {
		fmt.Println("error:", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("content-type", "application/json")
	w.Write([]byte(responsebody))

}

func main() {
	http.HandleFunc("/ping", getPing)
	http.HandleFunc("/books", getBooks)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
