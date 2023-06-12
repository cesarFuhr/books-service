package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Book struct {
	Name      string  `json:"name"`
	Price     float32 `json:"price"`
	Inventory int     `json:"inventory"`
}

var bookslist = []struct { //Struct created to handle data of the list of books
	Name      string  `json:"name"`
	Price     float32 `json:"price"`
	Inventory int     `json:"inventory"`
}{
	{"Book 1", 30.20, 2},
	{"Book 2", 20.30, 1},
	{"Book 3", 32.20, 5},
}

func Ping(w http.ResponseWriter, r *http.Request) {
	metodo := r.Method
	fmt.Printf("metodo: %v\n", metodo)
	if metodo == "GET" {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func Books(w http.ResponseWriter, r *http.Request) {
	metodo := r.Method
	fmt.Printf("metodo: %v\n", metodo)
	switch {
	case metodo == "GET":
		getBooks(w, r)
	case metodo == "POST":
		postBooks(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func postBooks(w http.ResponseWriter, r *http.Request) {
	//TO DO:
	//Read the Json body
	var newBook = Book{"sem nome", -1, -1}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&newBook)
	if err != nil {
		fmt.Println("error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Printf("newBook: %+v\n", newBook)
	fmt.Printf("bookslist: %+v", bookslist)

	//Verify if the entry is in a valid format
	switch {
	case newBook.Name == "sem nome":
		w.WriteHeader(http.StatusPartialContent)
		w.Write([]byte("Insira um nome para cadastrar o livro."))
	case newBook.Price == -1: //-1 means that no price was added
		w.WriteHeader(http.StatusPartialContent)
		w.Write([]byte("Insira um preço para cadastrar o livro."))
	case newBook.Inventory == -1: //-1 menas that no inventory was added
		w.WriteHeader(http.StatusPartialContent)
		w.Write([]byte("Insira a quantidade deste livro em estoque."))
	default:

	}

	//Verify if that book already exists in the database

	//Atribute an ID to the entry

	//Store the book in the database

	//Return a sucess message
}

func getBooks(w http.ResponseWriter, r *http.Request) {
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
	http.HandleFunc("/ping", Ping)
	http.HandleFunc("/books", Books)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
