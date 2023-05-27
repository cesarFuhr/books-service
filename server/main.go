package main

import (
	"log"
	"net/http"
)

func getPing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	//http.StatusMethodNotAllowed
}

func getBooks(w http.ResponseWriter, r *http.Request) {
	responsebody := `[
		{
		  "name": "Book 1",
		  "price": 30.20,
		  "invetory": 2
		},
		{
		  "name": "Book 2",
		  "price": 20.30,
		  "invetory": 1
		},
		{
		  "name": "Book 3",
		  "price": 32.20,
		  "invetory": 5
		},
	  ]`
	//io.WriteString(w, responsebody) Poderia ser com esse método também,
	// aí não seria necessário converter a string numa slice of bytes.
	w.Write([]byte(responsebody))
}

func main() {
	http.HandleFunc("/ping", getPing)
	http.HandleFunc("/books", getBooks)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
