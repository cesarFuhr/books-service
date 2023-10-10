package http

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/books-service/cmd/api/book"
	"github.com/books-service/cmd/api/database"
	bookerrors "github.com/books-service/cmd/api/errors"
	"github.com/google/uuid"
)

type BookHandler struct {
	bookService book.ServiceAPI
}

func NewBookHandler(bookService book.ServiceAPI) *BookHandler {
	return &BookHandler{bookService: bookService}
}

//==========HTTP ADDRESSERS:===========

/* Addresses a call to "/books/(expected id here)" according to the requested action.  */
func (h *BookHandler) bookById(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	switch method {
	case http.MethodGet:
		h.getBookById(w, r)
		return
	case http.MethodPut:
		updateBook(w, r)
		return
	case http.MethodDelete:
		h.archiveStatusBook(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

/* Addresses a call to "/books" according to the requested action.  */
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

//==========HTTP HANDLERS:===========

/* Change the status of a book to "archived". */
func (h *BookHandler) archiveStatusBook(w http.ResponseWriter, r *http.Request) {
	id, err := isolateId(w, r)
	if err != nil {
		return
	}

	archivedBook, err := h.bookService.ArchiveStatusBook(id)
	if err != nil {
		if errors.Is(err, bookerrors.ErrResponseBookNotFound) {
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

/* Returns the book with that specific ID. */
func (h *BookHandler) getBookById(w http.ResponseWriter, r *http.Request) {
	id, err := isolateId(w, r)
	if err != nil {
		return
	}
	//Searching for that ID on Book Service:
	returnedBook, err := h.bookService.GetBook(id)
	if err != nil {
		if errors.Is(err, bookerrors.ErrResponseBookNotFound) {
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

//==========AUXILIARY FUNCTIONS:===========

/* Isolates the ID from the URL. */
func isolateId(w http.ResponseWriter, r *http.Request) (id uuid.UUID, err error) {
	justId, _ := strings.CutPrefix(r.URL.Path, "/books/")
	id, err = uuid.Parse(justId)
	if err != nil {
		log.Println(err)
		responseJSON(w, http.StatusBadRequest, bookerrors.ErrResponseIdInvalidFormat)
		return id, err
	}
	return id, nil
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

/* Verifies if all fields are correctly filled and update the book in the db. */
func updateBook(w http.ResponseWriter, r *http.Request) {
	id, err := isolateId(w, r)
	if err != nil {
		return
	}

	var bookEntry book.Book
	err = json.NewDecoder(r.Body).Decode(&bookEntry) //Read the Json body and save the entry to bookEntry
	if err != nil {
		log.Println(err)
		errR := bookerrors.ErrResponse{
			Code:    bookerrors.ErrResponseBookEntryInvalidJSON.Code,
			Message: bookerrors.ErrResponseBookEntryInvalidJSON.Message + err.Error(),
		}
		responseJSON(w, http.StatusBadRequest, errR)
		return
	}

	err = book.FilledFields(bookEntry) //Verify if all entry fields are filled.
	if err != nil {
		responseJSON(w, http.StatusBadRequest, err)
		return
	}

	bookEntry.ID = id
	bookEntry.UpdatedAt = time.Now().UTC().Round(time.Millisecond) //Atribute a new updating time to the new entry.

	updatedBook, err := database.UpdateOnDB(bookEntry) //Update the book in the database
	if err != nil {
		if errors.Is(err, bookerrors.ErrResponseBookNotFound) {
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
	var bookEntry book.Book
	err := json.NewDecoder(r.Body).Decode(&bookEntry) //Read the Json body and save the entry to bookEntry
	if err != nil {
		log.Println(err)
		errR := bookerrors.ErrResponse{
			Code:    bookerrors.ErrResponseBookEntryInvalidJSON.Code,
			Message: bookerrors.ErrResponseBookEntryInvalidJSON.Message + err.Error(),
		}
		responseJSON(w, http.StatusBadRequest, errR)
		return
	}

	err = book.FilledFields(bookEntry) //Verify if all entry fields are filled.
	if err != nil {
		responseJSON(w, http.StatusBadRequest, err)
		return
	}

	bookEntry.ID = uuid.New() //Atribute an ID to the entry

	bookEntry.CreatedAt = time.Now().UTC().Round(time.Millisecond) //Atribute creating and updating time to the new entry. UpdateAt can change later.
	bookEntry.UpdatedAt = bookEntry.CreatedAt

	storedBook, err := database.StoreOnDB(bookEntry) //Store the book in the database
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
			responseJSON(w, http.StatusBadRequest, bookerrors.ErrResponseQueryPriceInvalidFormat)
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
			responseJSON(w, http.StatusBadRequest, bookerrors.ErrResponseQueryPriceInvalidFormat)
			return
		}
		maxPrice32 = float32(maxPrice64)
	} else {
		maxPrice32 = 9999.99 //max value to field price on db, set to: numeric(6,2)
	}

	sortBy, sortDirection, valid := extractOrderParams(query)
	if !valid {
		responseJSON(w, http.StatusBadRequest, bookerrors.ErrResponseQuerySortByInvalid)
		return
	}

	archived := false
	archivedStr := query.Get("archived")
	if archivedStr == "true" {
		archived = true
	}

	itemsTotal, err := database.CountRows(name, minPrice32, maxPrice32, archived)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if itemsTotal == 0 {
		responseJSON(w, http.StatusOK, []book.Book{})
		return
	}

	pagesTotal, page, pageSize, err := pagination(query, itemsTotal)
	if err != nil {
		responseJSON(w, http.StatusBadRequest, err)
		return
	}

	//Ask filtered list to db:
	returnedBooks, err := database.ListBooks(name, minPrice32, maxPrice32, sortBy, sortDirection, archived, page, pageSize)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	type PagedBooks struct {
		PageCurrent int         `json:"page_current"`
		PageTotal   int         `json:"page_total"`
		PageSize    int         `json:"page_size"`
		ItemsTotal  int         `json:"items_total"`
		Results     []book.Book `json:"results"`
	}

	pageOfBooksList := PagedBooks{
		PageCurrent: page,
		PageTotal:   pagesTotal,
		PageSize:    pageSize,
		ItemsTotal:  itemsTotal,
		Results:     returnedBooks,
	}

	responseJSON(w, http.StatusOK, pageOfBooksList)
}

/*Validates and prepares the pagination parameters of the query.*/
func pagination(query url.Values, itemsTotal int) (pagesTotal, page, pageSize int, err error) {

	pageStr := query.Get("page") //Convert page value to int and set default to 1.
	if pageStr == "" {
		page = 1
	} else {
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			return 0, 0, 0, bookerrors.ErrResponseQueryPageInvalid
		}
		if page <= 0 {
			return 0, 0, 0, bookerrors.ErrResponseQueryPageInvalid
		}
	}

	pageSizeStr := query.Get("page_size") //Convert page_size value to int and set default to 10.
	if pageSizeStr == "" {
		pageSize = 10
	} else {
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil {
			return 0, 0, 0, bookerrors.ErrResponseQueryPageInvalid
		}
		if !(0 < pageSize && pageSize < 31) {
			return 0, 0, 0, bookerrors.ErrResponseQueryPageInvalid
		}
	}

	pagesTotal = int(math.Ceil(float64(itemsTotal) / float64(pageSize)))
	if page > pagesTotal {
		return 0, 0, 0, bookerrors.ErrResponseQueryPageOutOfRange
	}

	return pagesTotal, page, pageSize, nil
}

/*Validates and prepares the ordering parameters of the query.*/
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