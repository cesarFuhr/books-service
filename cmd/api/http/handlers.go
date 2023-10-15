package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/books-service/cmd/api/book"
	"github.com/google/uuid"
)

type BookHandler struct {
	bookService book.ServiceAPI
}

func NewBookHandler(bookService book.ServiceAPI) *BookHandler {
	return &BookHandler{bookService: bookService}
}

/* Addresses a call to "/books/(expected id here)" according to the requested action.  */
func (h *BookHandler) bookById(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	switch method {
	case http.MethodGet:
		h.getBookById(w, r)
		return
	case http.MethodPut:
		h.updateBook(w, r)
		return
	case http.MethodDelete:
		h.archiveBook(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

/* Addresses a call to "/books" according to the requested action.  */
func (h *BookHandler) books(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	switch method {
	case http.MethodGet:
		h.listBooks(w, r)
		return
	case http.MethodPost:
		h.createBook(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

/* Change the status of a book to "archived". */
func (h *BookHandler) archiveBook(w http.ResponseWriter, r *http.Request) {
	id, err := isolateId(w, r)
	if err != nil {
		return
	}

	archivedBook, err := h.bookService.ArchiveBook(id)
	if err != nil {
		if errors.Is(err, book.ErrResponseBookNotFound) {
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

/* Validates the entry, then stores the entry as a new book. */
func (h *BookHandler) createBook(w http.ResponseWriter, r *http.Request) {
	var bookEntry book.Book
	err := json.NewDecoder(r.Body).Decode(&bookEntry) //Read the Json body and save the entry to bookEntry
	if err != nil {
		log.Println(err)
		errR := book.ErrResponse{
			Code:    book.ErrResponseBookEntryInvalidJSON.Code,
			Message: book.ErrResponseBookEntryInvalidJSON.Message + err.Error(),
		}
		responseJSON(w, http.StatusBadRequest, errR)
		return
	}

	err = book.FilledFields(bookEntry) //Verify if all entry fields are filled.
	if err != nil {
		responseJSON(w, http.StatusBadRequest, err)
		return
	}

	storedBook, err := h.bookService.CreateBook(bookEntry) //Store the book in the database
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	responseJSON(w, http.StatusCreated, storedBook)
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
		if errors.Is(err, book.ErrResponseBookNotFound) {
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

/* Returns a list of the stored books. */
func (h *BookHandler) listBooks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	name := query.Get("name")

	var minPrice32 float32
	minPriceStr := query.Get("min_price")
	if minPriceStr != "" {
		minPrice64, err := strconv.ParseFloat(minPriceStr, 32)
		if err != nil {
			responseJSON(w, http.StatusBadRequest, book.ErrResponseQueryPriceInvalidFormat)
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
			responseJSON(w, http.StatusBadRequest, book.ErrResponseQueryPriceInvalidFormat)
			return
		}
		maxPrice32 = float32(maxPrice64)
	} else {
		maxPrice32 = book.PriceMax
	}

	sortBy, sortDirection, valid := extractOrderParams(query)
	if !valid {
		responseJSON(w, http.StatusBadRequest, book.ErrResponseQuerySortByInvalid)
		return
	}

	archived := false
	archivedStr := query.Get("archived")
	if archivedStr == "true" {
		archived = true
	}

	page, pageSize, valid := extractPageParams(query)
	if !valid {
		responseJSON(w, http.StatusBadRequest, book.ErrResponseQueryPageInvalid)
		return
	}

	pagedBooks, err := h.bookService.ListBooks(name, minPrice32, maxPrice32, sortBy, sortDirection, archived, page, pageSize)
	if err != nil {
		if errors.Is(err, book.ErrResponseFromRespository) {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		responseJSON(w, http.StatusBadRequest, err)
		return
	}
	responseJSON(w, http.StatusOK, pagedBooks)
}

/* Validates the entry, then updates the asked book. */
func (h *BookHandler) updateBook(w http.ResponseWriter, r *http.Request) {
	id, err := isolateId(w, r)
	if err != nil {
		return
	}

	var bookEntry book.Book
	err = json.NewDecoder(r.Body).Decode(&bookEntry) //Read the Json body and save the entry to bookEntry
	if err != nil {
		log.Println(err)
		errR := book.ErrResponse{
			Code:    book.ErrResponseBookEntryInvalidJSON.Code,
			Message: book.ErrResponseBookEntryInvalidJSON.Message + err.Error(),
		}
		responseJSON(w, http.StatusBadRequest, errR)
		return
	}

	err = book.FilledFields(bookEntry) //Verify if all entry fields are filled.
	if err != nil {
		responseJSON(w, http.StatusBadRequest, err)
		return
	}

	updatedBook, err := h.bookService.UpdateBook(bookEntry, id) //Update the stored book
	if err != nil {
		if errors.Is(err, book.ErrResponseBookNotFound) {
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

/* Isolates the ID from the URL. */
func isolateId(w http.ResponseWriter, r *http.Request) (id uuid.UUID, err error) {
	justId, _ := strings.CutPrefix(r.URL.Path, "/books/")
	id, err = uuid.Parse(justId)
	if err != nil {
		log.Println(err)
		responseJSON(w, http.StatusBadRequest, book.ErrResponseIdInvalidFormat)
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

/*Validates and prepares the extractPageParams parameters of the query.*/
func extractPageParams(query url.Values) (page, pageSize int, valid bool) {
	var err error
	pageStr := query.Get("page") //Convert page value to int and set default to 1.
	if pageStr == "" {
		page = 1
	} else {
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			return 0, 0, false
		}
		if page <= 0 {
			return 0, 0, false
		}
	}

	pageSizeStr := query.Get("page_size") //Convert page_size value to int and set default to 10.
	if pageSizeStr == "" {
		pageSize = 10
	} else {
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil {
			return 0, 0, false
		}
		if !(0 < pageSize && pageSize < 31) {
			return 0, 0, false
		}
	}

	return page, pageSize, true
}
