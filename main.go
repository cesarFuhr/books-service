package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

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

//==============DATABASE FUNCTIONS:=================

/* Connects to the database trought a connection string and returns a pointer to a valid DB object (*sql.DB). */
func connectDb() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
	return db
}

/* Verifies if there is already a book with the name of "newBook" in the database. If yes, returns it. */
func SameNameOnDB(newBook Book) (unique bool, returnedBook Book) {
	sqlStatement := `SELECT id, name, price, inventory  FROM bookstable WHERE name=$1;`
	foundRow := dbObject.QueryRow(sqlStatement, newBook.Name)
	switch err := foundRow.Scan(&returnedBook.ID, &returnedBook.Name, &returnedBook.Price, &returnedBook.Inventory); err {
	case sql.ErrNoRows:
		unique = true
		return
	case nil:
		unique = false
		return
	default:
		panic(err)
	}
}

/* Stores the book into the database, checks and returns it if succeed. */
func StoreOnDB(newBook Book) (fail bool, storedBook Book) {
	sqlStatement := `
	INSERT INTO bookstable (id, name, price, inventory)
	VALUES ($1, $2, $3, $4)
	RETURNING *`
	createdRow := dbObject.QueryRow(sqlStatement, newBook.ID, newBook.Name, *newBook.Price, *newBook.Inventory)
	switch err := createdRow.Scan(&storedBook.ID, &storedBook.Name, &storedBook.Price, &storedBook.Inventory); err {
	case sql.ErrNoRows:
		fail = true
		return
	case nil:
		fail = false
		return
	default:
		panic(err)
	}
}

//==========HTTP COMMUNICATION FUNCTIONS:===========

/* Tests the http server connection.  */
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

/* Handles a call to /books and redirects depending on the requested action.  */
func Books(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	switch method {
	case http.MethodGet:
		getBooks(w, r)
		return
	case http.MethodPost:
		CreateBook(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

/* Verifies if all entry fields are filled and returns a warning message if so. */
func FilledFields(newBook Book) string {
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
	return blankFields
}

/* Stores the entry as a new book in the database, if there isn't one with the same name yet. */
func CreateBook(w http.ResponseWriter, r *http.Request) {

	var newBook Book
	err := json.NewDecoder(r.Body).Decode(&newBook) //Read the Json body and save the entry to newBook
	if err != nil {
		log.Println("error while decoding the json entry:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	blankFields := FilledFields(newBook) //Verify if all entry fields are filled.
	if blankFields != "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(blankFields))
		return
	}

	unique, returnedBook := SameNameOnDB(newBook) //Verify if the already there is a book with the same name in the database
	if !unique {
		sameNameBook, err := json.Marshal(returnedBook)
		if err != nil {
			log.Println("error while Marshalling returnedBook:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Este livro já existe na base de dados:\n" + string(sameNameBook)))
		return
	}

	newBook.ID = uuid.New() //Atribute an ID to the entry

	fail, storedBook := StoreOnDB(newBook) //Store the book in the database
	if fail {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Não foi possível adicionar o livro:\n"))
		return
	}
	showBook, err := json.Marshal(storedBook)
	if err != nil {
		log.Println("error while Marshalling storedBook:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Livro adicionado com sucesso:\n" + string(showBook)))
	return
}

//FIX IT CONSIDERING STORAGE ON DATABASE.
/* Return a list of the stored books. */
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
	//connect to db:
	dbObject = connectDb()
	defer dbObject.Close()

	//start http server:
	http.HandleFunc("/ping", Ping)
	http.HandleFunc("/books", Books)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
