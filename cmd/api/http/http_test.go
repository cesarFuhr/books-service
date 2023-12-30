package http_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/books-service/cmd/api/book"
	bookhttp "github.com/books-service/cmd/api/http"
	httpmock "github.com/books-service/cmd/api/http/mocks"
	"github.com/google/uuid"
	"github.com/matryer/is"
	"go.uber.org/mock/gomock"
)

func TestMain(m *testing.M) {
	var err error

	reqTimeoutStr := os.Getenv("HTTP_REQUEST_TIMEOUT") //This ENV must be written with a unit suffix, like seconds
	if reqTimeoutStr != "" {
		bookhttp.RequestTimeout, err = time.ParseDuration(reqTimeoutStr)
		if err != nil {
			log.Fatalln("getting request timeout from env: %w", err)
		}
	}

	os.Exit(m.Run())
}

func TestCreateBook(t *testing.T) {

	ctrl := gomock.NewController(t)
	mockAPI := httpmock.NewMockServiceAPI(ctrl)
	bookHandler := bookhttp.NewBookHandler(mockAPI)

	server := bookhttp.NewServer(bookhttp.ServerConfig{Port: 8080}, bookHandler)

	t.Run("creates a book without errors", func(t *testing.T) {
		is := is.New(t)

		reqBook := book.CreateBookRequest{
			Name:      "HTTP tester book",
			Price:     toPointer(float32(100.0)),
			Inventory: toPointer(99),
		}
		bookToCreate := `{
			"name": "HTTP tester book",
			"price": 100,
			"inventory": 99
		}`
		newID := uuid.New()
		expectedReturn := book.Book{
			ID:        newID,
			Name:      reqBook.Name,
			Price:     reqBook.Price,
			Inventory: reqBook.Inventory,
			CreatedAt: time.Now().UTC().Round(time.Millisecond),
			UpdatedAt: time.Now().UTC().Round(time.Millisecond),
			Archived:  false,
		}
		expectedJSONresponse := fmt.Sprintf(`{"id":"%s","name":"HTTP tester book","price":100,"inventory":99,"archived":false}`+"\n", newID)

		request, _ := http.NewRequest(http.MethodPost, "/books", strings.NewReader(bookToCreate))
		response := httptest.NewRecorder()

		mockAPI.EXPECT().CreateBook(gomock.Any(), reqBook).Return(expectedReturn, nil) //Maybe we should change Any() for the right context, defined as global at the top

		server.Handler.ServeHTTP(response, request)

		body, _ := io.ReadAll(response.Result().Body)

		is.True(response.Result().StatusCode == 201)
		is.Equal(string(body), expectedJSONresponse)

	})

	t.Run("expected invalid json error", func(t *testing.T) {
		is := is.New(t)

		invalidBookToCreate := `{
				"name": "test with missing coma after price",
				"price": 100
				"inventory": 99
			}`
		expectedJSONresponse := fmt.Sprintln(`{"error_code":102,"error_message":"invalid json request.invalid character '\"' after object key:value pair"}`)

		request, _ := http.NewRequest(http.MethodPost, "/books", strings.NewReader(invalidBookToCreate))
		response := httptest.NewRecorder()

		server.Handler.ServeHTTP(response, request)

		body, _ := io.ReadAll(response.Result().Body)

		is.True(response.Result().StatusCode == 400)
		is.Equal(string(body), expectedJSONresponse)

	})

	t.Run("expected blank fields error", func(t *testing.T) {
		is := is.New(t)

		invalidBookToCreate := `{
			"name": "test with missing inventory",
			"price": 100
		}`
		expectedJSONresponse := fmt.Sprintln(`{"error_code":100,"error_message":"all the fields - name, price and inventory - must be filled correctly."}`)

		request, _ := http.NewRequest(http.MethodPost, "/books", strings.NewReader(invalidBookToCreate))
		response := httptest.NewRecorder()

		server.Handler.ServeHTTP(response, request)

		body, _ := io.ReadAll(response.Result().Body)

		is.True(response.Result().StatusCode == 400)
		is.Equal(string(body), expectedJSONresponse)

	})

	t.Run("expected context timeout error", func(t *testing.T) {
		is := is.New(t)

		reqBook := book.CreateBookRequest{
			Name:      "HTTP tester book",
			Price:     toPointer(float32(100.0)),
			Inventory: toPointer(99),
		}
		bookToCreate := `{
			"name": "HTTP tester book",
			"price": 100,
			"inventory": 99
		}`
		newID := uuid.New()
		expectedReturn := book.Book{
			ID:        newID,
			Name:      reqBook.Name,
			Price:     reqBook.Price,
			Inventory: reqBook.Inventory,
			CreatedAt: time.Now().UTC().Round(time.Millisecond),
			UpdatedAt: time.Now().UTC().Round(time.Millisecond),
			Archived:  false,
		}
		expectedJSONresponse := fmt.Sprintln(`{"error_code":109,"error_message":"error from context:context deadline exceeded"}`)

		ctxTest, cancel := context.WithTimeout(context.Background(), bookhttp.RequestTimeout)
		defer cancel()

		request, _ := http.NewRequestWithContext(ctxTest, http.MethodPost, "/books", strings.NewReader(bookToCreate))
		response := httptest.NewRecorder()

		mockAPI.EXPECT().CreateBook(gomock.Any(), reqBook).DoAndReturn(func(ctx context.Context, req book.CreateBookRequest) (book.Book, error) {
			time.Sleep(bookhttp.RequestTimeout + time.Second)
			log.Println("context error: ", ctxTest.Err())
			return expectedReturn, ctxTest.Err()
		})

		server.Handler.ServeHTTP(response, request)

		body, _ := io.ReadAll(response.Result().Body)

		is.True(response.Result().StatusCode == 504)
		is.Equal(string(body), expectedJSONresponse)

	})

	t.Run("expected context canceled error", func(t *testing.T) {
		is := is.New(t)

		reqBook := book.CreateBookRequest{
			Name:      "HTTP tester book",
			Price:     toPointer(float32(100.0)),
			Inventory: toPointer(99),
		}
		bookToCreate := `{
			"name": "HTTP tester book",
			"price": 100,
			"inventory": 99
		}`
		newID := uuid.New()
		expectedReturn := book.Book{
			ID:        newID,
			Name:      reqBook.Name,
			Price:     reqBook.Price,
			Inventory: reqBook.Inventory,
			CreatedAt: time.Now().UTC().Round(time.Millisecond),
			UpdatedAt: time.Now().UTC().Round(time.Millisecond),
			Archived:  false,
		}
		expectedJSONresponse := fmt.Sprintln(`{"error_code":109,"error_message":"error from context:context canceled"}`)

		ctxTest, cancel := context.WithTimeout(context.Background(), bookhttp.RequestTimeout)
		defer cancel()

		request, _ := http.NewRequestWithContext(ctxTest, http.MethodPost, "/books", strings.NewReader(bookToCreate))
		response := httptest.NewRecorder()

		mockAPI.EXPECT().CreateBook(gomock.Any(), reqBook).DoAndReturn(func(ctx context.Context, req book.CreateBookRequest) (book.Book, error) {
			cancel()
			log.Println("context error: ", ctxTest.Err())
			return expectedReturn, ctxTest.Err()
		})

		server.Handler.ServeHTTP(response, request)

		body, _ := io.ReadAll(response.Result().Body)

		is.True(response.Result().StatusCode == 504)

		is.Equal(string(body), expectedJSONresponse)

	})
}
func TestListBooks(t *testing.T) {

	ctrl := gomock.NewController(t)
	mockAPI := httpmock.NewMockServiceAPI(ctrl)
	bookHandler := bookhttp.NewBookHandler(mockAPI)

	server := bookhttp.NewServer(bookhttp.ServerConfig{Port: 8080}, bookHandler)

	// Setting up, creating books to be listed.
	var testBookslist []book.Book
	listSize := 30
	for i := 0; i < listSize; i++ {
		b := book.Book{
			ID:        uuid.New(),
			Name:      fmt.Sprintf("Book number %06v", i),
			Price:     toPointer(float32((i * 100) + 1)),
			Inventory: toPointer(100 - i),
			CreatedAt: time.Now().UTC().Round(time.Millisecond),
			UpdatedAt: time.Now().UTC().Round(time.Millisecond),
			Archived:  false,
		}
		testBookslist = append(testBookslist, b)
	}

	t.Run("lists all books with defalult values, without errors", func(t *testing.T) {
		is := is.New(t)

		// Setting query parameters
		params := book.ListBooksRequest{
			Name:          "",
			MinPrice:      0,
			MaxPrice:      book.PriceMax,
			SortBy:        "name",
			SortDirection: "asc",
			Archived:      false,
			Page:          1,
			PageSize:      10,
		}
		url := "/books"

		expectedReturn := book.PagedBooks{
			PageCurrent: params.Page,
			PageTotal:   3,
			PageSize:    params.PageSize,
			ItemsTotal:  listSize,
			Results:     testBookslist[0:10],
		}

		expectedJSONresponse, err := json.Marshal(pagedBooksToResponse(expectedReturn))
		is.NoErr(err)
		expectedJSONresponse = append(expectedJSONresponse, []byte("\n")...)

		request, _ := http.NewRequest(http.MethodGet, url, nil)
		response := httptest.NewRecorder()

		mockAPI.EXPECT().ListBooks(params).Return(expectedReturn, nil)

		server.Handler.ServeHTTP(response, request)

		body, _ := io.ReadAll(response.Result().Body)

		is.True(response.Result().StatusCode == 200)
		is.Equal(string(body), string(expectedJSONresponse))

	})
	t.Run("lists all books with not defalult values, without errors", func(t *testing.T) {
		is := is.New(t)

		// Setting query parameters
		params := book.ListBooksRequest{
			Name:          "Book",
			MinPrice:      1,
			MaxPrice:      10000.01,
			SortBy:        "inventory",
			SortDirection: "desc",
			Archived:      true,
			Page:          2,
			PageSize:      15,
		}
		url := "/books?name=Book&max_price=10000.01&min_price=1&sort_by=inventory&sort_direction=desc&page_size=15&page=2&archived=true"

		expectedReturn := book.PagedBooks{
			PageCurrent: params.Page,
			PageTotal:   2,
			PageSize:    params.PageSize,
			ItemsTotal:  listSize,
			Results:     testBookslist[15:30],
		}

		expectedJSONresponse, err := json.Marshal(pagedBooksToResponse(expectedReturn))
		is.NoErr(err)
		expectedJSONresponse = append(expectedJSONresponse, []byte("\n")...)

		request, _ := http.NewRequest(http.MethodGet, url, nil)
		response := httptest.NewRecorder()

		mockAPI.EXPECT().ListBooks(params).Return(expectedReturn, nil)

		server.Handler.ServeHTTP(response, request)

		body, _ := io.ReadAll(response.Result().Body)

		is.True(response.Result().StatusCode == 200)
		is.Equal(string(body), string(expectedJSONresponse))

	})

	t.Run("expected order parameters error", func(t *testing.T) {
		is := is.New(t)

		url := "/books?sort_by=invalid string"

		expectedJSONresponse, err := json.Marshal(book.ErrResponseQuerySortByInvalid)
		is.NoErr(err)
		expectedJSONresponse = append(expectedJSONresponse, []byte("\n")...)

		request, _ := http.NewRequest(http.MethodGet, url, nil)
		response := httptest.NewRecorder()

		server.Handler.ServeHTTP(response, request)

		body, _ := io.ReadAll(response.Result().Body)

		is.True(response.Result().StatusCode == 400)
		is.Equal(string(body), string(expectedJSONresponse))

	})

	t.Run("expected page parameters error", func(t *testing.T) {
		is := is.New(t)

		url := "/books?page_size=35"

		expectedJSONresponse, err := json.Marshal(book.ErrResponseQueryPageInvalid)
		is.NoErr(err)
		expectedJSONresponse = append(expectedJSONresponse, []byte("\n")...)

		request, _ := http.NewRequest(http.MethodGet, url, nil)
		response := httptest.NewRecorder()

		server.Handler.ServeHTTP(response, request)

		body, _ := io.ReadAll(response.Result().Body)

		is.True(response.Result().StatusCode == 400)
		is.Equal(string(body), string(expectedJSONresponse))

	})
}

func toPointer[T any](v T) *T {
	return &v
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
