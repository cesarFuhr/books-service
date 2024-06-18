package book_test

import (
	"context"
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

		mockRepo.EXPECT().CreateOrder(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, o book.Order) (book.Order, error) {
			is.True(o.Order_ID != uuid.Nil)
			is.True(o.Purchaser_ID == someUser)
			is.True(o.Order_status == "accepting_items")
			is.True(o.CreatedAt.Compare(time.Now().Round(time.Millisecond)) <= 0)
			is.True(o.UpdatedAt.Compare(time.Now().Round(time.Millisecond)) <= 0)
			return o, nil
		})

		newOrder, err := mS.CreateOrder(ctx, someUser)
		is.NoErr(err)
		is.True(newOrder.Order_ID != uuid.Nil)
		is.True(newOrder.Purchaser_ID == someUser)
		is.True(newOrder.Order_status == "accepting_items")
		is.True(newOrder.CreatedAt.Compare(time.Now().Round(time.Millisecond)) <= 0)
		is.True(newOrder.UpdatedAt.Compare(time.Now().Round(time.Millisecond)) <= 0)
	})
}
