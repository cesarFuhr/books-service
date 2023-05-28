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
	Livros := []struct {
		name      string
		price     float32
		inventory int
	}{
		{"Book 1", 30.20, 2},
		{"Book 2", 20.30, 1},
		{"Book 3", 32.20, 5},
	}

	stringLivros := make([]string, len(Livros))

	var responsebody string

	for i, v := range Livros {
		stringLivros[i] = fmt.Sprintln("name:", v.name, "\nprice:", v.price, "\ninventory:", v.inventory, "\n")
		responsebody = fmt.Sprint(responsebody, stringLivros[i])
	}

	/*responsebody := `[
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
	  ]` */
	//io.WriteString(w, responsebody) Poderia ser com esse método também,
	// aí não seria necessário converter a string numa slice of bytes.
	w.Write([]byte(responsebody))
}

func main() {
	http.HandleFunc("/ping", getPing)
	http.HandleFunc("/books", getBooks)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
