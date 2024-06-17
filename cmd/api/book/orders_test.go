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

func TestCreateOrder(t *testing.T) {

	t.Run("creates an order with a generic user", func(t *testing.T) {

		is := is.New(t)
		ctrl := gomock.NewController(t)
		mockRepo := bookmock.NewMockRepository(ctrl)
		mockNtfy := bookmock.NewMockNotifier(ctrl)
		mS := book.NewService(mockRepo, mockNtfy, notificationsTimeout)

		someUser := uuid.New()

		/*
			mockRepo.EXPECT().CreateOrder(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, o book.Order) (book.Order, error) {
				is.True(o.Order_ID != uuid.Nil)
				is.True(o.Purchaser_ID == someUser)
				is.True(o.Order_status == "accepting_items")
				is.True(o.CreatedAt.Compare(time.Now().Round(time.Millisecond)) <= 0)
				is.True(o.UpdatedAt.Compare(time.Now().Round(time.Millisecond)) <= 0)
				return o, nil
			})
		*/

		/*	reqBook := book.CreateBookRequest{
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

			wg := sync.WaitGroup{}
			wg.Add(1)
			b := book.Book{}
			mockNtfy.EXPECT().BookCreated(gomock.Any(), gomock.AssignableToTypeOf(b)).DoAndReturn(func(_ context.Context, _ book.Book) error {
				defer wg.Done()
				return nil
			})

			createdBook, err := mS.CreateBook(ctx, reqBook)
			is.NoErr(err)
			is.True(createdBook.ID != uuid.Nil)
			is.Equal(createdBook.Name, reqBook.Name)
			is.Equal(createdBook.Price, reqBook.Price)
			is.Equal(createdBook.Inventory, reqBook.Inventory)
			is.True(!createdBook.Archived)

			wg.Wait()
		*/

		newOrder, err := mS.CreateOrder(ctx, someUser)
		is.NoErr(err)
		is.True(newOrder.Order_ID != uuid.Nil)
		is.True(newOrder.Purchaser_ID == someUser)
		is.True(newOrder.Order_status == "accepting_items")
		is.True(newOrder.CreatedAt.Compare(time.Now().Round(time.Millisecond)) <= 0)
		is.True(newOrder.UpdatedAt.Compare(time.Now().Round(time.Millisecond)) <= 0)
	})
}
