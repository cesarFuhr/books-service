package book_test

import (
	"fmt"
	"testing"

	"github.com/books-service/cmd/api/book"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

var s *book.Service

//var mS *MockRepositoryMockRecorder

/*func TestMain(m *testing.M) {

	os.Exit(m.Run())
}*/

func TestCreateBook(t *testing.T) {
	t.Run("creates a book without errors", func(t *testing.T) {
		is := is.New(t)
		reqBook := book.CreateBookRequest{
			Name:      "Service tester book",
			Price:     toPointer(float32(100.0)),
			Inventory: toPointer(99),
		}
		fmt.Println(reqBook)
		createdBook, err := s.CreateBook(reqBook)
		is.NoErr(err)
		is.True(createdBook.ID != uuid.Nil)
		is.Equal(createdBook.Name, reqBook.Name)
		is.Equal(createdBook.Price, reqBook.Price)
		is.Equal(createdBook.Inventory, reqBook.Inventory)
		is.True(!createdBook.Archived)
	})
}

func toPointer[T any](v T) *T {
	return &v
}

/*func TestCreateBook(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)
}*/
