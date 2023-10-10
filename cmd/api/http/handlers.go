package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/books-service/cmd/api/book"
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
		h.updateBook(w, r)
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

/* Validates the entry, then stores the entry as a new book. */
func (h *BookHandler) createBook(w http.ResponseWriter, r *http.Request) {
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

/* Returns a list of the stored books. */
func (h *BookHandler) listBooks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	pagedBooks, err := h.bookService.ListBooks(query)
	if err != nil {
		if errors.Is(err, bookerrors.ErrResponseFromRespository) {
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

	updatedBook, err := h.bookService.UpdateBook(bookEntry, id) //Update the stored book
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
