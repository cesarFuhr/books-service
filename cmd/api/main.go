package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Book struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Price     *float32  `json:"price"`
	Inventory *int      `json:"inventory"`
}

var dbObjectGlobal *sql.DB

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

/* Handles a call to /books/(exected id here) and redirects depending on the requested action.  */
func bookById(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	if method == http.MethodGet {
		getBookById(w, r)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

/* Returns the book with that specific ID. */
func getBookById(w http.ResponseWriter, r *http.Request) {
	//Isolating ID:
	justId, _ := strings.CutPrefix(r.URL.Path, "/books/")
	id, err := uuid.Parse(justId)
	if err != nil {
		log.Println(err)
		responseJSON(w, http.StatusBadRequest, errResponseIdInvalidFormat)
		return
	}
	//Searching for that ID on database:
	returnedBook, err := searchById(id)
	if err != nil {
		if errors.Is(err, errBookNotFound) {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	responseJSON(w, http.StatusOK, returnedBook)
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

/*Writes a JSON response into a http.ResponseWriter. */
func responseJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

/* Stores the entry as a new book in the database, if there isn't one with the same name yet. */
func createBook(w http.ResponseWriter, r *http.Request) {

	var newBook Book
	err := json.NewDecoder(r.Body).Decode(&newBook) //Read the Json body and save the entry to newBook
	if err != nil {
		log.Println(err)
		errR := errResponse{
			Code:    errResponseCreateBookInvalidJSON.Code,
			Message: errResponseCreateBookInvalidJSON.Message + err.Error(),
		}
		responseJSON(w, http.StatusBadRequest, errR)
		return
	}

	filled := filledFields(newBook) //Verify if all entry fields are filled.
	if !filled {
		responseJSON(w, http.StatusBadRequest, errResponseCreateBookBlankFileds)
		return
	}

	unique, err := sameNameOnDB(newBook) //Verify if the already there is a book with the same name in the database
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !unique {
		responseJSON(w, http.StatusBadRequest, errResponseCreateBookNameConflict)
		return
	}

	newBook.ID = uuid.New() //Atribute an ID to the entry

	storedBook, err := storeOnDB(newBook) //Store the book in the database
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	responseJSON(w, http.StatusCreated, storedBook)
}

/* Returns a list of the stored books. */
func getBooks(w http.ResponseWriter, r *http.Request) {
	returnedBooks, err := listBooks()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	responseJSON(w, http.StatusOK, returnedBooks)
}

func main() {
	//connect to db:
	dbObject, err := connectDb()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	dbObjectGlobal = dbObject
	defer dbObjectGlobal.Close()

	//apply migrations:
	err = migrationUp()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	//start http server:
	http.HandleFunc("/ping", ping)
	http.HandleFunc("/books", books)
	http.HandleFunc("/books/", bookById)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
