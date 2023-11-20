package book_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/books-service/cmd/api/book"
	bookmock "github.com/books-service/cmd/api/book/mocks"
	"github.com/google/uuid"
	"github.com/matryer/is"
	gomock "go.uber.org/mock/gomock"
)

var mS *book.Service

/*func TestMain(m *testing.M) {

	os.Exit(m.Run())
}*/

func TestCreateBook(t *testing.T) {
	t.Run("creates a book without errors", func(t *testing.T) {
		is := is.New(t)
		ctrl := gomock.NewController(t)
		mockRepo := bookmock.NewMockRepository(ctrl)
		mS = book.NewService(mockRepo)

		reqBook := book.CreateBookRequest{
			Name:      "Service tester book",
			Price:     toPointer(float32(100.0)),
			Inventory: toPointer(99),
		}
		fmt.Println(reqBook)

		mockRepo.EXPECT().CreateBook(book.Book{
			ID:        uuid.New(), //Atribute an ID to the entry
			Name:      "Service tester book",
			Price:     toPointer(float32(100.0)),
			Inventory: toPointer(99),
			CreatedAt: time.Now().UTC().Round(time.Millisecond),
			UpdatedAt: time.Now().UTC().Round(time.Millisecond),
			Archived:  false,
		})
		createdBook, err := mS.CreateBook(reqBook)
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

/*func TestCreateBookCOMMOCK(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)
	mS := book.NewService(mockRepo)

}*/
