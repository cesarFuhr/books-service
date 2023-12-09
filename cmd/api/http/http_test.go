package http_test

import (
	"net/http"
	"net/http/httptest"
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

//func TestMain(m *testing.M) {
//	os.Exit(m.Run())
//}

func TestCreateBook(t *testing.T) {

	ctrl := gomock.NewController(t)
	mockAPI := httpmock.NewMockServiceAPI(ctrl)
	bookHandler := bookhttp.NewBookHandler(mockAPI)

	//server := bookhttp.NewServer(bookhttp.ServerConfig{Port: 8080}, bookHandler)
	t.Run("creates a book without errors", func(t *testing.T) {
		is := is.New(t)

		reqBook := book.CreateBookRequest{ //FORMAT THIS STRING LATER!!!
			Name:      "HTTP tester book",
			Price:     toPointer(float32(100.0)),
			Inventory: toPointer(99),
		}
		bookToCreate := `{
			"name": "HTTP tester book",
			"price": 100,
			"inventory": 99
		}`
		expectedReturn := book.Book{
			ID:        uuid.New(),
			Name:      reqBook.Name,
			Price:     reqBook.Price,
			Inventory: reqBook.Inventory,
			CreatedAt: time.Now().UTC().Round(time.Millisecond),
			UpdatedAt: time.Now().UTC().Round(time.Millisecond),
			Archived:  false,
		}

		request, _ := http.NewRequest("", "", strings.NewReader(bookToCreate)) //HTTP method and url are empty here, proving that the route isn't being tested.
		response := httptest.NewRecorder()

		mockAPI.EXPECT().CreateBook(reqBook).Return(expectedReturn, nil)

		bookHandler.EXPOcreateBook(response, request) //To call the handle function directly it must be exported!

		//server.Handler.ServeHTTP(response, request)

		is.True(response.Result().StatusCode == 201)

	})
}

func toPointer[T any](v T) *T {
	return &v
}
