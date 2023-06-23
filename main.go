package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	//"strings"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

const (
	host     = "db"
	port     = 5432
	user     = "postgres"
	password = "chevas"
	dbname   = "booksdb"
)

type Book struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Price     *float32  `json:"price"`
	Inventory *int      `json:"inventory"`
}

var bookslist []Book
var dbObject *sql.DB

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
		log.Println("error while decoding the json entry:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Verify if that book already exists in the database
	sqlStatementQuery := `SELECT id, name, price, inventory  FROM bookstable WHERE name=$1;`
	var returnedBook Book
	foundRow := dbObject.QueryRow(sqlStatementQuery, newBook.Name)
	switch err := foundRow.Scan(&returnedBook.ID, &returnedBook.Name, &returnedBook.Price, &returnedBook.Inventory); err {
	case sql.ErrNoRows:
		break
	case nil:
		showBook, err := json.Marshal(returnedBook)
		if err != nil {
			log.Println("error while Marshalling returnedBook:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Este livro já existe na base de dados:\n" + string(showBook)))
		return
	default:
		panic(err)
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
	sqlStatementCreate := `
INSERT INTO bookstable (id, name, price, inventory)
VALUES ($1, $2, $3, $4)
RETURNING *`

	var storedBook Book
	createdRow := dbObject.QueryRow(sqlStatementCreate, newBook.ID, newBook.Name, *newBook.Price, *newBook.Inventory)

	//Check and Return a sucess message
	switch err := createdRow.Scan(&storedBook.ID, &storedBook.Name, &storedBook.Price, &storedBook.Inventory); err {
	case sql.ErrNoRows:
		log.Println("No rows were returned!")
		return
	case nil:
		showBook, err := json.Marshal(storedBook)
		if err != nil {
			log.Println("error while Marshalling storedBook:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Livro adicionado com sucesso:\n" + string(showBook)))
		return
	default:
		panic(err)
	}

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
	//set database:
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	dbObject = db
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")

	//start http server:
	http.HandleFunc("/ping", Ping)
	http.HandleFunc("/books", Books)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
