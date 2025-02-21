package http

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/books-service/cmd/api/book"
	"github.com/google/uuid"
)

type BookHandler struct {
	bookService    book.ServiceAPI
	requestTimeout time.Duration
}

func NewBookHandler(bookService book.ServiceAPI, reqTimeout time.Duration) *BookHandler {
	return &BookHandler{
		bookService:    bookService,
		requestTimeout: reqTimeout,
	}
}

/* Addresses a call to "/books/(expected id here)" according to the requested action.  */
func (h *BookHandler) bookById(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(h.requestTimeout))
	defer cancel()
	r = r.WithContext(ctx)

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

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(h.requestTimeout))
	defer cancel()
	r = r.WithContext(ctx)

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

	archivedBook, err := h.bookService.ArchiveBook(r.Context(), id)
	if err != nil {
		handleError(err, w, r)
		return
	}

	responseJSON(w, http.StatusOK, bookToResponse(archivedBook))
}

type BookEntry struct {
	Name      string   `json:"name"`
	Price     *float32 `json:"price"`
	Inventory *int     `json:"inventory"`
}

/* Validates the entry, then stores the entry as a new book. */
func (h *BookHandler) createBook(w http.ResponseWriter, r *http.Request) {
	var bookEntry BookEntry
	err := json.NewDecoder(r.Body).Decode(&bookEntry) //Read the Json body and save the entry to bookEntry
	if err != nil {
		log.Println(err)
		errR := book.ErrResponse{
			Code:    book.ErrResponseEntryInvalidJSON.Code,
			Message: book.ErrResponseEntryInvalidJSON.Message + err.Error(),
		}
		responseJSON(w, http.StatusBadRequest, errR)
		return
	}

	err = FilledBookFields(bookEntry) //Verify if all entry fields are filled.
	if err != nil {
		responseJSON(w, http.StatusBadRequest, err)
		return
	}

	reqBook := bookToCreateReq(bookEntry)

	storedBook, err := h.bookService.CreateBook(r.Context(), reqBook)
	if err != nil {
		handleError(err, w, r)
		return
	}

	responseJSON(w, http.StatusCreated, bookToResponse(storedBook))
}

/* Validates the entry, then updates the asked book. */
func (h *BookHandler) updateBook(w http.ResponseWriter, r *http.Request) {
	id, err := isolateId(w, r)
	if err != nil {
		return
	}

	var bookEntry BookEntry
	err = json.NewDecoder(r.Body).Decode(&bookEntry) //Read the Json body and save the entry to bookEntry
	if err != nil {
		log.Println(err)
		errR := book.ErrResponse{
			Code:    book.ErrResponseEntryInvalidJSON.Code,
			Message: book.ErrResponseEntryInvalidJSON.Message + err.Error(),
		}
		responseJSON(w, http.StatusBadRequest, errR)
		return
	}

	err = FilledBookFields(bookEntry) //Verify if all entry fields are filled.
	if err != nil {
		responseJSON(w, http.StatusBadRequest, err)
		return
	}

	reqBook := bookToUpdateReq(bookEntry, id)

	updatedBook, err := h.bookService.UpdateBook(r.Context(), reqBook) //Update the stored book
	if err != nil {
		handleError(err, w, r)
		return
	}

	responseJSON(w, http.StatusOK, bookToResponse(updatedBook))
}

/* Returns the book with that specific ID. */
func (h *BookHandler) getBookById(w http.ResponseWriter, r *http.Request) {
	id, err := isolateId(w, r)
	if err != nil {
		return
	}
	//Searching for that ID on Book Service:
	returnedBook, err := h.bookService.GetBook(r.Context(), id)
	if err != nil {
		handleError(err, w, r)
		return
	}

	responseJSON(w, http.StatusOK, bookToResponse(returnedBook))
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

	params := book.ListBooksRequest{
		Name:          name,
		MinPrice:      minPrice32,
		MaxPrice:      maxPrice32,
		SortBy:        sortBy,
		SortDirection: sortDirection,
		Archived:      archived,
		Page:          page,
		PageSize:      pageSize,
	}

	pagedBooks, err := h.bookService.ListBooks(r.Context(), params)
	if err != nil {
		handleError(err, w, r)
		return
	}

	responseJSON(w, http.StatusOK, pagedBooksToResponse(pagedBooks))
}

/* Verifies if all Book entry fields are filled and returns a warning message if so. */
func FilledBookFields(bookEntry BookEntry) error {
	if bookEntry.Name == "" {
		return book.ErrResponseBookEntryBlankFields
	}
	if bookEntry.Price == nil {
		return book.ErrResponseBookEntryBlankFields
	}
	if bookEntry.Inventory == nil {
		return book.ErrResponseBookEntryBlankFields
	}

	return nil
}

/* Converts from BookEntry type to CreateBookRequest type, with no json tags. */
func bookToCreateReq(b BookEntry) book.CreateBookRequest {
	return book.CreateBookRequest{
		Name:      b.Name,
		Price:     b.Price,
		Inventory: b.Inventory,
	}
}

/* Converts from BookEntry type to UpdateBookRequest type, with no json tags. */
func bookToUpdateReq(b BookEntry, id uuid.UUID) book.UpdateBookRequest {
	return book.UpdateBookRequest{
		ID:        id,
		Name:      b.Name,
		Price:     b.Price,
		Inventory: b.Inventory,
	}
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

type BookResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Price     *float32  `json:"price"`
	Inventory *int      `json:"inventory"`
	Archived  bool      `json:"archived"`
}

/*Copy the fields of a book object to an http layer struct with json tags*/
func bookToResponse(b book.Book) BookResponse {
	return BookResponse{
		ID:        b.ID,
		Name:      b.Name,
		Price:     b.Price,
		Inventory: b.Inventory,
		Archived:  b.Archived,
	}
}

type PageOfBooksResponse struct {
	PageCurrent int            `json:"page_current"`
	PageTotal   int            `json:"page_total"`
	PageSize    int            `json:"page_size"`
	ItemsTotal  int            `json:"items_total"`
	Results     []BookResponse `json:"results"`
}

/*Copy the fields of a PagedBooks object to an http layer struct with json tags*/
func pagedBooksToResponse(page book.PagedBooks) PageOfBooksResponse {
	results := []BookResponse{}
	for _, book := range page.Results {
		results = append(results, bookToResponse(book))
	}

	return PageOfBooksResponse{
		PageCurrent: page.PageCurrent,
		PageTotal:   page.PageTotal,
		PageSize:    page.PageSize,
		ItemsTotal:  page.ItemsTotal,
		Results:     results,
	}
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

func handleError(err error, w http.ResponseWriter, r *http.Request) {
	log.Println(err)
	if errors.As(err, &book.ErrResponse{}) {
		switch {
		case errors.Is(err, book.ErrResponseQueryPageOutOfRange):
			responseJSON(w, http.StatusBadRequest, book.ErrResponseQueryPageOutOfRange)
			return
		case errors.Is(err, book.ErrResponseBookNotFound):
			responseJSON(w, http.StatusNotFound, book.ErrResponseBookNotFound)
			return
		case errors.Is(err, book.ErrResponseBookIsArchived):
			responseJSON(w, http.StatusBadRequest, book.ErrResponseBookIsArchived)
			return
		case errors.Is(err, book.ErrResponseInsufficientInventory):
			responseJSON(w, http.StatusBadRequest, book.ErrResponseInsufficientInventory)
			return
		case errors.Is(err, book.ErrResponseOrderNotFound):
			responseJSON(w, http.StatusNotFound, book.ErrResponseOrderNotFound)
			return
		case errors.Is(err, book.ErrResponseOrderNotAcceptingItems):
			responseJSON(w, http.StatusBadRequest, book.ErrResponseOrderNotAcceptingItems)
			return
		case errors.Is(err, book.ErrResponseBookNotAtOrder):
			responseJSON(w, http.StatusBadRequest, book.ErrResponseBookNotAtOrder)
			return
		}
	} else if errors.Is(err, context.DeadlineExceeded) {
		responseJSON(w, http.StatusGatewayTimeout, book.ErrResponseRequestTimeout)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
}
