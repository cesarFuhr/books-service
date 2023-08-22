package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Book struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Price     *float32  `json:"price"`
	Inventory *int      `json:"inventory"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
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

	newBook.CreatedAt = time.Now().UTC().Round(time.Millisecond) //Atribute creating and updating time to the new entry. UpdateAt can change later.
	newBook.UpdatedAt = newBook.CreatedAt

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

	//Extract and adapt query params:
	//FILTERING PARAMS:
	query := r.URL.Query()

	name := query.Get("name")

	var minPrice32 float32
	minPriceStr := query.Get("min_price")
	if minPriceStr != "" {
		minPrice64, err := strconv.ParseFloat(minPriceStr, 32)
		if err != nil {
			responseJSON(w, http.StatusBadRequest, errResponseQueryPriceInvalidFormat)
			return
		}
		minPrice32 = float32(minPrice64)
	} else {
		minPrice32 = 0
	}

	var maxPrice32 float32
	maxPriceStr := query.Get("max_price")
	if maxPriceStr != "" {
		maxPrice64, err := strconv.ParseFloat(maxPriceStr, 32)
		if err != nil {
			responseJSON(w, http.StatusBadRequest, errResponseQueryPriceInvalidFormat)
			return
		}
		maxPrice32 = float32(maxPrice64)
	} else {
		maxPrice32 = 9999.99 //max value to field price on db, set to: numeric(6,2)
	}

	sortBy, valid := extractOrderParmams(query)
	if !valid {
		responseJSON(w, http.StatusBadRequest, errResponseQuerySortByInvalid)
		return
	}

	//Ask filtered list to db:
	returnedBooks, err := listBooks(name, minPrice32, maxPrice32, sortBy)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	responseJSON(w, http.StatusOK, returnedBooks)
}

func extractOrderParmams(query url.Values) (sortBy string, valid bool) {
	sortBy = query.Get("sort_by")
	switch sortBy {
	case "":
		sortBy = "name"
	case "name":
		break
	case "price":
		break
	case "inventory":
		break
	case "created_at":
		break
	case "updated_at":
		break //IMPLEMENT THIS LATER, WHIT FUNCTION UPDATE BOOK
		//https://x-team.com/blog/automatic-timestamps-with-postgresql/
		//https://www.postgresqltutorial.com/postgresql-date-functions/postgresql-current_timestamp/
		//https://www.postgresql.org/docs/15/sql-createtrigger.html
	default:
		return sortBy, false
	}

	return sortBy, true
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
