package http_test

import (
	"fmt"
	"io"
	"log"
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

	server := bookhttp.NewServer(bookhttp.ServerConfig{Port: 8080}, bookHandler)
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
		expectedJSONresponse := fmt.Sprintf(`{"id":"%s","name":"HTTP tester book","price":100,"inventory":99,"archived":false}`, newID)

		request, _ := http.NewRequest(http.MethodPost, "/books", strings.NewReader(bookToCreate))
		response := httptest.NewRecorder()

		mockAPI.EXPECT().CreateBook(reqBook).Return(expectedReturn, nil)

		server.Handler.ServeHTTP(response, request)

		body, _ := io.ReadAll(response.Result().Body)

		log.Printf("%s, %T \n", string(body), string(body))
		log.Printf("%s, %T \n", expectedJSONresponse, expectedJSONresponse)
		log.Println("comparision is:", strings.Compare(string(body), expectedJSONresponse))

		is.True(response.Result().StatusCode == 201)
		is.Equal(string(body), expectedJSONresponse)

	})
}

func toPointer[T any](v T) *T {
	return &v
}
