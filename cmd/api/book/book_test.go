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

func TestCreateBook(t *testing.T) {
	t.Run("creates a book without errors", func(t *testing.T) {
		is := is.New(t)
		ctrl := gomock.NewController(t)
		mockRepo := bookmock.NewMockRepository(ctrl)
		mS := book.NewService(mockRepo)

		reqBook := book.CreateBookRequest{
			Name:      "Service tester book",
			Price:     toPointer(float32(100.0)),
			Inventory: toPointer(99),
		}
		fmt.Println(reqBook)

		mockRepo.EXPECT().CreateBook(gomock.Any()).DoAndReturn(func(b book.Book) (book.Book, error) {
			is.True(b.ID != uuid.Nil)
			is.Equal(b.Name, reqBook.Name)
			is.Equal(b.Price, reqBook.Price)
			is.Equal(b.Inventory, reqBook.Inventory)
			is.True(!b.Archived)
			is.True(b.CreatedAt.Compare(time.Now().Round(time.Millisecond)) <= 0)
			is.True(b.CreatedAt.Compare(time.Now().Round(time.Millisecond)) <= 0)
			return b, nil
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
