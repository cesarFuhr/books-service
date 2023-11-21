package book_test

import (
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

		mockRepo.EXPECT().CreateBook(gomock.Any()).DoAndReturn(func(b book.Book) (book.Book, error) {
			is.True(b.ID != uuid.Nil)
			is.Equal(b.Name, reqBook.Name)
			is.Equal(b.Price, reqBook.Price)
			is.Equal(b.Inventory, reqBook.Inventory)
			is.True(!b.Archived)
			is.True(b.CreatedAt.Compare(time.Now().Round(time.Millisecond)) <= 0)
			is.True(b.UpdatedAt.Compare(time.Now().Round(time.Millisecond)) <= 0)
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

func TestUpdateBook(t *testing.T) {
	t.Run("updates a book without errors", func(t *testing.T) {
		is := is.New(t)
		ctrl := gomock.NewController(t)
		mockRepo := bookmock.NewMockRepository(ctrl)
		mS := book.NewService(mockRepo)

		reqBook := book.UpdateBookRequest{
			ID:        uuid.New(),
			Name:      "Updated service tester book",
			Price:     toPointer(float32(100.0)),
			Inventory: toPointer(99),
		}

		mockRepo.EXPECT().UpdateBook(gomock.Any()).DoAndReturn(func(b book.Book) (book.Book, error) {
			is.Equal(b.ID, reqBook.ID)
			is.Equal(b.Name, reqBook.Name)
			is.Equal(b.Price, reqBook.Price)
			is.Equal(b.Inventory, reqBook.Inventory)
			is.True(b.UpdatedAt.Compare(time.Now().Round(time.Millisecond)) <= 0)
			return b, nil
		})

		updatedBook, err := mS.UpdateBook(reqBook)
		is.NoErr(err)
		is.Equal(updatedBook.ID, reqBook.ID)
		is.Equal(updatedBook.Name, reqBook.Name)
		is.Equal(updatedBook.Price, reqBook.Price)
		is.Equal(updatedBook.Inventory, reqBook.Inventory)
		is.True(updatedBook.UpdatedAt.Compare(updatedBook.CreatedAt.Round(time.Millisecond)) > 0)
	})
}

func TestArchiveBook(t *testing.T) {
	t.Run("archives a book without errors", func(t *testing.T) {
		is := is.New(t)
		ctrl := gomock.NewController(t)
		mockRepo := bookmock.NewMockRepository(ctrl)
		mS := book.NewService(mockRepo)

		id := uuid.New()

		mockRepo.EXPECT().SetBookArchiveStatus(id, true)

		_, err := mS.ArchiveBook(id)
		is.NoErr(err)
	})
}

func TestGetBook(t *testing.T) {
	t.Run("Gets a book by ID without errors", func(t *testing.T) {
		is := is.New(t)
		ctrl := gomock.NewController(t)
		mockRepo := bookmock.NewMockRepository(ctrl)
		mS := book.NewService(mockRepo)

		id := uuid.New()

		mockRepo.EXPECT().GetBookByID(id)

		_, err := mS.GetBook(id)
		is.NoErr(err)
	})
}

/*
func TestListBooks(t *testing.T) {
	t.Run("list books without errors", func(t *testing.T) {
		is := is.New(t)
		ctrl := gomock.NewController(t)
		mockRepo := bookmock.NewMockRepository(ctrl)
		mS := book.NewService(mockRepo)

		/* reqBooks := book.ListBooksRequest{
					Name:      "",
					MinPrice      float32
					MaxPrice      float32
					SortBy        string
					SortDirection string
					Archived      bool
					Page          int
					PageSize      int
		} */
/*
		itemsTotal := 30

		mockRepo.EXPECT().ListBooksTotals(reqBooks.Name, reqBooks.MinPrice, reqBooks.MaxPrice, reqBooks.Archived).Return(itemsTotal, nil)

		updatedBook, err := mS.ListBooks(reqBook)
		is.NoErr(err)
		is.Equal(updatedBook.ID, reqBook.ID)
		is.Equal(updatedBook.Name, reqBook.Name)
		is.Equal(updatedBook.Price, reqBook.Price)
		is.Equal(updatedBook.Inventory, reqBook.Inventory)
		is.True(updatedBook.UpdatedAt.Compare(updatedBook.CreatedAt.Round(time.Millisecond)) > 0)
	})
} */

func toPointer[T any](v T) *T {
	return &v
}
