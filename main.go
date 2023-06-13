package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type Book struct {
	Name      string    `json:"name"`
	Price     float32   `json:"price"`
	Inventory int       `json:"inventory"`
	ID        uuid.UUID `json:"id"`
}

var bookslist = []struct { //Struct created to handle data of the list of books
	Name      string    `json:"name"`
	Price     float32   `json:"price"`
	Inventory int       `json:"inventory"`
	ID        uuid.UUID `json:"id"`
}{
	{"Book 1", 30.20, 2, uuid.New()},
	{"Book 2", 20.30, 1, uuid.New()},
	{"Book 3", 32.20, 5, uuid.New()},
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
	var newBook = Book{"sem nome", -1, -1, uuid.Nil}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&newBook)
	if err != nil {
		fmt.Println("error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Printf("newBook: %+v\n", newBook)
	fmt.Printf("bookslist: %+v\n", bookslist)

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
	for i := range bookslist {
		bookAlreadyExists := strings.EqualFold(bookslist[i].Name, newBook.Name)
		if bookAlreadyExists {
			warning := fmt.Sprintf("Este livro já existe na base de dados: %+v", bookslist[i])
			w.Write([]byte(warning))
			return
		}
	}

	//Atribute an ID to the entry
	newBook.ID = uuid.New()

	//Store the book in the database
	bookslist = append(bookslist, newBook)
	fmt.Printf("bookslist: %+v\n", bookslist)

	//Return a sucess message
	w.Write([]byte("Livro adicionado com sucesso"))
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
