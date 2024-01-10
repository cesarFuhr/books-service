package book_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/books-service/cmd/api/book"
	bookmock "github.com/books-service/cmd/api/book/mocks"
	"github.com/books-service/cmd/api/notifications"
	"github.com/google/uuid"
	"github.com/matryer/is"
	gomock "go.uber.org/mock/gomock"
)

var ctx context.Context = context.Background()

var ntfy *notifications.Ntfy
var notificationsTimeout = 1 * time.Second

func TestMain(m *testing.M) {
	//temporary copied from main.go:
	notificationsBaseURL := "someURL"
	enableNotifications := true
	ntfy = notifications.NewNtfy(enableNotifications, notificationsBaseURL)

	os.Exit(m.Run())
}

func TestCreateBook(t *testing.T) {

	t.Run("creates a book without errors", func(t *testing.T) {
		is := is.New(t)
		ctrl := gomock.NewController(t)
		mockRepo := bookmock.NewMockRepository(ctrl)
		mS := book.NewService(mockRepo, ntfy, notificationsTimeout)

		reqBook := book.CreateBookRequest{
			Name:      "Service tester book",
			Price:     toPointer(float32(100.0)),
			Inventory: toPointer(99),
		}

		mockRepo.EXPECT().CreateBook(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, b book.Book) (book.Book, error) {
			is.True(b.ID != uuid.Nil)
			is.Equal(b.Name, reqBook.Name)
			is.Equal(b.Price, reqBook.Price)
			is.Equal(b.Inventory, reqBook.Inventory)
			is.True(!b.Archived)
			is.True(b.CreatedAt.Compare(time.Now().Round(time.Millisecond)) <= 0)
			is.True(b.UpdatedAt.Compare(time.Now().Round(time.Millisecond)) <= 0)
			return b, nil
		})

		createdBook, err := mS.CreateBook(ctx, reqBook)
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
		mS := book.NewService(mockRepo, ntfy, notificationsTimeout)

		reqBook := book.UpdateBookRequest{
			ID:        uuid.New(),
			Name:      "Updated service tester book",
			Price:     toPointer(float32(100.0)),
			Inventory: toPointer(99),
		}

		mockRepo.EXPECT().UpdateBook(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, b book.Book) (book.Book, error) {
			is.Equal(b.ID, reqBook.ID)
			is.Equal(b.Name, reqBook.Name)
			is.Equal(b.Price, reqBook.Price)
			is.Equal(b.Inventory, reqBook.Inventory)
			is.True(b.UpdatedAt.Compare(time.Now().Round(time.Millisecond)) <= 0)
			return b, nil
		})

		updatedBook, err := mS.UpdateBook(ctx, reqBook)
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
		mS := book.NewService(mockRepo, ntfy, notificationsTimeout)

		id := uuid.New()

		mockRepo.EXPECT().SetBookArchiveStatus(gomock.Any(), id, true)

		_, err := mS.ArchiveBook(ctx, id)
		is.NoErr(err)
	})
}

func TestGetBook(t *testing.T) {
	t.Run("Gets a book by ID without errors", func(t *testing.T) {
		is := is.New(t)
		ctrl := gomock.NewController(t)
		mockRepo := bookmock.NewMockRepository(ctrl)
		mS := book.NewService(mockRepo, ntfy, notificationsTimeout)

		id := uuid.New()

		mockRepo.EXPECT().GetBookByID(gomock.Any(), id)

		_, err := mS.GetBook(ctx, id)
		is.NoErr(err)
	})
}

func TestListBooks(t *testing.T) {
	is := is.New(t)
	ctrl := gomock.NewController(t)
	mockRepo := bookmock.NewMockRepository(ctrl)
	mS := book.NewService(mockRepo, ntfy, notificationsTimeout)
	t.Run("list first page of stored books without errors, paginated with exact division", func(t *testing.T) {
		//Setting specific subtest values:
		reqBooks := book.ListBooksRequest{
			Name:          "",
			MinPrice:      0.0,
			MaxPrice:      book.PriceMax,
			SortBy:        "name",
			SortDirection: "asc",
			Archived:      true,
			Page:          1,
			PageSize:      10,
		}
		itemsTotal := 30
		expectedPagesTotal := 3 //(itemsTotal / PageSize) up rounded to next integer
		results := []book.Book{}
		//-------------------------------

		mockRepo.EXPECT().ListBooksTotals(gomock.Any(), reqBooks.Name, reqBooks.MinPrice, reqBooks.MaxPrice, reqBooks.Archived).Return(itemsTotal, nil)
		mockRepo.EXPECT().ListBooks(gomock.Any(), reqBooks.Name, reqBooks.MinPrice, reqBooks.MaxPrice, reqBooks.SortBy, reqBooks.SortDirection, reqBooks.Archived, reqBooks.Page, reqBooks.PageSize).Return(results, nil)

		pageOfBooksList, err := mS.ListBooks(ctx, reqBooks)
		is.NoErr(err)
		is.Equal(pageOfBooksList.PageCurrent, reqBooks.Page)
		is.Equal(pageOfBooksList.PageTotal, expectedPagesTotal)
		is.Equal(pageOfBooksList.PageSize, reqBooks.PageSize)
		is.Equal(pageOfBooksList.ItemsTotal, itemsTotal)
		is.Equal(pageOfBooksList.Results, results)
	})

	t.Run("list first page of stored books without errors, paginated with not exact division", func(t *testing.T) {
		//Setting specific subtest values:
		reqBooks := book.ListBooksRequest{
			Name:          "",
			MinPrice:      0.0,
			MaxPrice:      book.PriceMax,
			SortBy:        "name",
			SortDirection: "asc",
			Archived:      true,
			Page:          1,
			PageSize:      10,
		}
		itemsTotal := 31
		expectedPagesTotal := 4 //(itemsTotal / PageSize) up rounded to next integer
		results := []book.Book{}
		//-------------------------------

		mockRepo.EXPECT().ListBooksTotals(gomock.Any(), reqBooks.Name, reqBooks.MinPrice, reqBooks.MaxPrice, reqBooks.Archived).Return(itemsTotal, nil)
		mockRepo.EXPECT().ListBooks(gomock.Any(), reqBooks.Name, reqBooks.MinPrice, reqBooks.MaxPrice, reqBooks.SortBy, reqBooks.SortDirection, reqBooks.Archived, reqBooks.Page, reqBooks.PageSize).Return(results, nil)

		pageOfBooksList, err := mS.ListBooks(ctx, reqBooks)
		is.NoErr(err)
		is.Equal(pageOfBooksList.PageCurrent, reqBooks.Page)
		is.Equal(pageOfBooksList.PageTotal, expectedPagesTotal)
		is.Equal(pageOfBooksList.PageSize, reqBooks.PageSize)
		is.Equal(pageOfBooksList.ItemsTotal, itemsTotal)
		is.Equal(pageOfBooksList.Results, results)
	})

	t.Run("list second page of books without errors", func(t *testing.T) {
		//Setting specific subtest values:
		reqBooks := book.ListBooksRequest{
			Name:          "",
			MinPrice:      0.0,
			MaxPrice:      book.PriceMax,
			SortBy:        "name",
			SortDirection: "asc",
			Archived:      true,
			Page:          2,
			PageSize:      10,
		}
		itemsTotal := 30
		expectedPagesTotal := 3 //(itemsTotal / PageSize) up rounded to next integer
		results := []book.Book{}
		//-------------------------------

		mockRepo.EXPECT().ListBooksTotals(gomock.Any(), reqBooks.Name, reqBooks.MinPrice, reqBooks.MaxPrice, reqBooks.Archived).Return(itemsTotal, nil)
		mockRepo.EXPECT().ListBooks(gomock.Any(), reqBooks.Name, reqBooks.MinPrice, reqBooks.MaxPrice, reqBooks.SortBy, reqBooks.SortDirection, reqBooks.Archived, reqBooks.Page, reqBooks.PageSize).Return(results, nil)

		pageOfBooksList, err := mS.ListBooks(ctx, reqBooks)
		is.NoErr(err)
		is.Equal(pageOfBooksList.PageCurrent, reqBooks.Page)
		is.Equal(pageOfBooksList.PageTotal, expectedPagesTotal)
		is.Equal(pageOfBooksList.PageSize, reqBooks.PageSize)
		is.Equal(pageOfBooksList.ItemsTotal, itemsTotal)
		is.Equal(pageOfBooksList.Results, results)
	})

	t.Run("list books asking page out of range", func(t *testing.T) {
		//Setting specific subtest values:
		reqBooks := book.ListBooksRequest{
			Name:          "",
			MinPrice:      0.0,
			MaxPrice:      book.PriceMax,
			SortBy:        "name",
			SortDirection: "asc",
			Archived:      true,
			Page:          4,
			PageSize:      10,
		}
		itemsTotal := 30
		//expectedPagesTotal := 3 //(itemsTotal / PageSize) up rounded to next integer
		//results := []book.Book{}
		//-------------------------------

		mockRepo.EXPECT().ListBooksTotals(gomock.Any(), reqBooks.Name, reqBooks.MinPrice, reqBooks.MaxPrice, reqBooks.Archived).Return(itemsTotal, nil)
		//Its expected that the method returns before calling ListBooks due to the pagination error.
		//mockRepo.EXPECT().ListBooks(reqBooks.Name, reqBooks.MinPrice, reqBooks.MaxPrice, reqBooks.SortBy, reqBooks.SortDirection, reqBooks.Archived, reqBooks.Page, reqBooks.PageSize).Return(results, nil)

		pageOfBooksList, err := mS.ListBooks(ctx, reqBooks)
		is.True(errors.Is(err, book.ErrResponseQueryPageOutOfRange))
		is.Equal(pageOfBooksList, book.PagedBooks{})
	})

	t.Run("no books to list", func(t *testing.T) {
		//Setting specific subtest values:
		reqBooks := book.ListBooksRequest{
			Name:          "",
			MinPrice:      0.0,
			MaxPrice:      book.PriceMax,
			SortBy:        "name",
			SortDirection: "asc",
			Archived:      true,
			Page:          2,
			PageSize:      10,
		}
		itemsTotal := 0
		expectedPagesTotal := 0 //(itemsTotal / PageSize) up rounded to next integer
		results := []book.Book{}
		//-------------------------------

		mockRepo.EXPECT().ListBooksTotals(gomock.Any(), reqBooks.Name, reqBooks.MinPrice, reqBooks.MaxPrice, reqBooks.Archived).Return(itemsTotal, nil)
		//Its expected that the method returns before calling ListBooks since there is no books to list.
		//mockRepo.EXPECT().ListBooks(reqBooks.Name, reqBooks.MinPrice, reqBooks.MaxPrice, reqBooks.SortBy, reqBooks.SortDirection, reqBooks.Archived, reqBooks.Page, reqBooks.PageSize).Return(results, nil)

		pageOfBooksList, err := mS.ListBooks(ctx, reqBooks)
		is.NoErr(err)
		is.Equal(pageOfBooksList.PageCurrent, 0)
		is.Equal(pageOfBooksList.PageTotal, expectedPagesTotal)
		is.Equal(pageOfBooksList.PageSize, 0)
		is.Equal(pageOfBooksList.ItemsTotal, itemsTotal)
		is.Equal(pageOfBooksList.Results, results)
	})

	t.Run("expected error from database", func(t *testing.T) {
		//Setting specific subtest values:
		reqBooks := book.ListBooksRequest{
			Name:          "",
			MinPrice:      0.0,
			MaxPrice:      book.PriceMax,
			SortBy:        "name",
			SortDirection: "asc",
			Archived:      true,
			Page:          1,
			PageSize:      10,
		}
		itemsTotal := 30
		//expectedPagesTotal := 3 //(itemsTotal / PageSize) up rounded to next integer
		results := []book.Book{}
		dbErr := errors.New("fake error from database")
		errRepo := book.ErrResponse{
			Code:    book.ErrResponseFromRespository.Code,
			Message: book.ErrResponseFromRespository.Message + dbErr.Error(),
		}
		//-------------------------------

		mockRepo.EXPECT().ListBooksTotals(gomock.Any(), reqBooks.Name, reqBooks.MinPrice, reqBooks.MaxPrice, reqBooks.Archived).Return(itemsTotal, nil)

		mockRepo.EXPECT().ListBooks(gomock.Any(), reqBooks.Name, reqBooks.MinPrice, reqBooks.MaxPrice, reqBooks.SortBy, reqBooks.SortDirection, reqBooks.Archived, reqBooks.Page, reqBooks.PageSize).Return(results, dbErr)

		pageOfBooksList, err := mS.ListBooks(ctx, reqBooks)
		is.Equal(pageOfBooksList, book.PagedBooks{})
		is.Equal(err, errRepo)
	})

	t.Run("expected context timeout error", func(t *testing.T) {
		//Setting specific subtest values:
		reqBooks := book.ListBooksRequest{
			Name:          "",
			MinPrice:      0.0,
			MaxPrice:      book.PriceMax,
			SortBy:        "name",
			SortDirection: "asc",
			Archived:      true,
			Page:          1,
			PageSize:      10,
		}
		itemsTotal := 30
		results := []book.Book{}

		mockRepo.EXPECT().ListBooksTotals(gomock.Any(), reqBooks.Name, reqBooks.MinPrice, reqBooks.MaxPrice, reqBooks.Archived).Return(itemsTotal, nil)

		mockRepo.EXPECT().ListBooks(gomock.Any(), reqBooks.Name, reqBooks.MinPrice, reqBooks.MaxPrice, reqBooks.SortBy, reqBooks.SortDirection, reqBooks.Archived, reqBooks.Page, reqBooks.PageSize).Return(results, context.DeadlineExceeded)

		pageOfBooksList, err := mS.ListBooks(ctx, reqBooks)
		is.Equal(pageOfBooksList, book.PagedBooks{})
		is.Equal(err.Error(), "timeout on call to ListBooks: "+context.DeadlineExceeded.Error())
	})
}

func toPointer[T any](v T) *T {
	return &v
}
