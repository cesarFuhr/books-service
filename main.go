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
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Price     *float32  `json:"price"`
	Inventory *int      `json:"inventory"`
}

var bookslist []Book

func Ping(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	if method == http.MethodGet {
		w.WriteHeader(http.StatusNoContent)
		return
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func Books(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	switch method {
	case http.MethodGet:
		getBooks(w, r)
		return
	case http.MethodPost:
		postBooks(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func postBooks(w http.ResponseWriter, r *http.Request) {
	//Read the Json body and save the entry to newBook
	var newBook Book
	err := json.NewDecoder(r.Body).Decode(&newBook)
	if err != nil {
		log.Println("error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Verify if that book already exists in the database
	for i := range bookslist {
		bookAlreadyExists := strings.EqualFold(bookslist[i].Name, newBook.Name)
		if bookAlreadyExists {
			showBook, err := json.Marshal(bookslist[i])
			if err != nil {
				log.Println("error:", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			warning := fmt.Sprint("Este livro já existe na base de dados:\n" + string(showBook))
			w.Write([]byte(warning))
			return
		}
	}

	//Verify if the entry is in a valid format
	blankFields := ""
	if newBook.Name == "" {
		blankFields += "Insira um nome para cadastrar o livro.\n"
	}
	if newBook.Price == nil {
		blankFields += "Insira um preço para cadastrar o livro.\n"
	}
	if newBook.Inventory == nil {
		blankFields += "Insira a quantidade deste livro em estoque.\n"
	}
	if blankFields != "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(blankFields))
		return
	}

	//Atribute an ID to the entry
	newBook.ID = uuid.New()

	//Store the book in the database
	bookslist = append(bookslist, newBook)

	//Return a sucess message
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Livro adicionado com sucesso."))
}

func getBooks(w http.ResponseWriter, r *http.Request) {
	//Encoding in JSON to send through the Writer:
	responsebody, err := json.Marshal(bookslist)
	if err != nil {
		log.Println("error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write([]byte(responsebody))
}

func main() {
	http.HandleFunc("/ping", Ping)
	http.HandleFunc("/books", Books)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
