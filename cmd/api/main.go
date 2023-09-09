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
	Archived  bool      `json:"archived"`
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
	switch method {
	case http.MethodGet:
		getBookById(w, r)
		return
	case http.MethodPut:
		updateBook(w, r)
		return
	case http.MethodDelete:
		archiveStatusBook(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

/* Isolates the ID from the URL. */
func isolateId(w http.ResponseWriter, r *http.Request) (id uuid.UUID, err error) {
	justId, _ := strings.CutPrefix(r.URL.Path, "/books/")
	id, err = uuid.Parse(justId)
	if err != nil {
		log.Println(err)
		responseJSON(w, http.StatusBadRequest, errResponseIdInvalidFormat)
		return id, err
	}
	return id, nil
}

/* Returns the book with that specific ID. */
func getBookById(w http.ResponseWriter, r *http.Request) {
	id, err := isolateId(w, r)
	if err != nil {
		return
	}
	//Searching for that ID on database:
	returnedBook, err := searchById(id)
	if err != nil {
		if errors.Is(err, errResponseBookNotFound) {
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
func filledFields(bookEntry Book) error {
	if bookEntry.Name == "" {
		return errResponseBookEntryBlankFileds
	}
	if bookEntry.Price == nil {
		return errResponseBookEntryBlankFileds
	}
	if bookEntry.Inventory == nil {
		return errResponseBookEntryBlankFileds
	}

	return nil
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

/* Set 'true' to column 'archived' in the database */
func archiveStatusBook(w http.ResponseWriter, r *http.Request) {
	id, err := isolateId(w, r)
	if err != nil {
		return
	}
	archived := true //THIS FUNCTION IS READY TO BE IMPROVED TO ALLOW RESTORING BOOKS FROM ARCHIVE. JUST FIX THE ROUTING!
	archivedBook, err := archiveStatusOnDB(id, archived)
	if err != nil {
		if errors.Is(err, errResponseBookNotFound) {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	responseJSON(w, http.StatusOK, archivedBook)
}

/* Verifies if all fields are correctly filled and update the book in the db. */
func updateBook(w http.ResponseWriter, r *http.Request) {
	id, err := isolateId(w, r)
	if err != nil {
		return
	}

	var bookEntry Book
	err = json.NewDecoder(r.Body).Decode(&bookEntry) //Read the Json body and save the entry to bookEntry
	if err != nil {
		log.Println(err)
		errR := errResponse{
			Code:    errResponseBookEntryInvalidJSON.Code,
			Message: errResponseBookEntryInvalidJSON.Message + err.Error(),
		}
		responseJSON(w, http.StatusBadRequest, errR)
		return
	}

	err = filledFields(bookEntry) //Verify if all entry fields are filled.
	if err != nil {
		responseJSON(w, http.StatusBadRequest, err)
		return
	}

	bookEntry.ID = id
	bookEntry.UpdatedAt = time.Now().UTC().Round(time.Millisecond) //Atribute a new updating time to the new entry.

	updatedBook, err := updateOnDB(bookEntry) //Update the book in the database
	if err != nil {
		if errors.Is(err, errResponseBookNotFound) {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	responseJSON(w, http.StatusOK, updatedBook)
}

/* Stores the entry as a new book in the database. */
func createBook(w http.ResponseWriter, r *http.Request) {
	var bookEntry Book
	err := json.NewDecoder(r.Body).Decode(&bookEntry) //Read the Json body and save the entry to bookEntry
	if err != nil {
		log.Println(err)
		errR := errResponse{
			Code:    errResponseBookEntryInvalidJSON.Code,
			Message: errResponseBookEntryInvalidJSON.Message + err.Error(),
		}
		responseJSON(w, http.StatusBadRequest, errR)
		return
	}

	err = filledFields(bookEntry) //Verify if all entry fields are filled.
	if err != nil {
		responseJSON(w, http.StatusBadRequest, err)
		return
	}

	bookEntry.ID = uuid.New() //Atribute an ID to the entry

	bookEntry.CreatedAt = time.Now().UTC().Round(time.Millisecond) //Atribute creating and updating time to the new entry. UpdateAt can change later.
	bookEntry.UpdatedAt = bookEntry.CreatedAt

	storedBook, err := storeOnDB(bookEntry) //Store the book in the database
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

	sortBy, sortDirection, valid := extractOrderParams(query)
	if !valid {
		responseJSON(w, http.StatusBadRequest, errResponseQuerySortByInvalid)
		return
	}

	archived := false
	archivedStr := query.Get("archived")
	if archivedStr == "true" {
		archived = true
	}

	//Ask filtered list to db:
	returnedBooks, err := listBooks(name, minPrice32, maxPrice32, sortBy, sortDirection, archived)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	responseJSON(w, http.StatusOK, returnedBooks)
}

func extractOrderParams(query url.Values) (sortBy string, sortDirection string, valid bool) {
	sortDirection = query.Get("sort_direction")
	switch sortDirection {
	case "":
		sortDirection = "asc"
	case "asc":
		break
	case "desc":
		break
	default:
		return sortBy, sortDirection, false
	}

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
		break
	default:
		return sortBy, sortDirection, false
	}

	return sortBy, sortDirection, true
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
