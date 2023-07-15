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
	host     = "db" //outside container: "localhost"; inside container: "db"
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
type errorResponse struct {
	Code    int    `json:"error_code"`
	Message string `json:"error_message"`
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
func sameNameOnDB(newBook Book) (unique bool, returnedBook Book) {
	sqlStatement := `SELECT id, name, price, inventory  FROM bookstable WHERE name=$1;`
	foundRow := dbObject.QueryRow(sqlStatement, newBook.Name)
	var bookToReturn Book
	switch err := foundRow.Scan(&bookToReturn.ID, &bookToReturn.Name, &bookToReturn.Price, &bookToReturn.Inventory); err {
	case sql.ErrNoRows:
		return true, Book{}
	case nil:
		return false, bookToReturn
	default:
		panic(err)
	}
}

/* Stores the book into the database, checks and returns it if succeed. */
func storeOnDB(newBook Book) (fail bool, storedBook Book) {
	sqlStatement := `
	INSERT INTO bookstable (id, name, price, inventory)
	VALUES ($1, $2, $3, $4)
	RETURNING *`
	createdRow := dbObject.QueryRow(sqlStatement, newBook.ID, newBook.Name, *newBook.Price, *newBook.Inventory)
	var bookToReturn Book
	switch err := createdRow.Scan(&bookToReturn.ID, &bookToReturn.Name, &bookToReturn.Price, &bookToReturn.Inventory); err {
	case sql.ErrNoRows:
		return true, Book{}
	case nil:
		return false, bookToReturn
	default:
		panic(err)
	}
}

//==========HTTP COMMUNICATION FUNCTIONS:===========

/* Tests the http server connection.  */
func ping(w http.ResponseWriter, r *http.Request) {
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
func books(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	switch method {
	case http.MethodGet:
		getBooks(w, r)
		return
	case http.MethodPost:
		createBook(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

/* Verifies if all entry fields are filled and returns a warning message if so. */
func filledFields(newBook Book) bool {
	var filled bool
	if newBook.Name == "" {
		filled = false
		return filled
	}
	if newBook.Price == nil {
		filled = false
		return filled
	}
	if newBook.Inventory == nil {
		filled = false
		return filled
	}
	filled = true
	return filled
}

/* Stores the entry as a new book in the database, if there isn't one with the same name yet. */
func createBook(w http.ResponseWriter, r *http.Request) {

	var newBook Book
	err := json.NewDecoder(r.Body).Decode(&newBook) //Read the Json body and save the entry to newBook
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	filled := filledFields(newBook) //Verify if all entry fields are filled.
	if !filled {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(errResonseCreateBookBadInput)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}

	unique, _ /*returnedBook*/ := sameNameOnDB(newBook) //Verify if the already there is a book with the same name in the database
	if !unique {
		/*
			Keeping this block until we decide what to do with "returnedBook"

			sameNameBook, err := json.Marshal(returnedBook)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}	*/
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(errResonseCreateBookNameConflict)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		//	w.Write(append(error, sameNameBook...))
		return
	}

	newBook.ID = uuid.New() //Atribute an ID to the entry

	fail, storedBook := storeOnDB(newBook) //Store the book in the database
	if fail {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(storedBook)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
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
	http.HandleFunc("/ping", ping)
	http.HandleFunc("/books", books)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
