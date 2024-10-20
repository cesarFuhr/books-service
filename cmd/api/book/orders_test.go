package book_test

import (
	"context"
	"errors"
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
			is.True(o.OrderID != uuid.Nil)
			is.Equal(o.PurchaserID, someUser)
			is.True(o.OrderStatus == "accepting_items")
			is.True(o.CreatedAt.Compare(time.Now().Round(time.Millisecond)) <= 0)
			is.True(o.UpdatedAt.Compare(time.Now().Round(time.Millisecond)) <= 0)
			return o, nil
		})

		newOrder, err := mS.CreateOrder(ctx, someUser)
		is.NoErr(err)
		is.True(newOrder.OrderID != uuid.Nil)
		is.Equal(newOrder.PurchaserID, someUser)
		is.True(newOrder.OrderStatus == "accepting_items")
		is.True(newOrder.CreatedAt.Compare(time.Now().Round(time.Millisecond)) <= 0)
		is.True(newOrder.UpdatedAt.Compare(time.Now().Round(time.Millisecond)) <= 0)
	})
}

func TestListOrderItems(t *testing.T) {
	t.Run("list items of an order, without errors", func(t *testing.T) {
		is := is.New(t)
		ctrl := gomock.NewController(t)
		mockRepo := bookmock.NewMockRepository(ctrl)
		mockNtfy := bookmock.NewMockNotifier(ctrl)
		mS := book.NewService(mockRepo, mockNtfy, notificationsTimeout)

		newOrderID := uuid.New()

		mockRepo.EXPECT().ListOrderItems(gomock.Any(), newOrderID).Return(book.Order{OrderID: newOrderID}, nil)

		order, err := mS.ListOrderItems(ctx, newOrderID)
		is.NoErr(err)
		is.Equal(order.OrderID, newOrderID)
		is.Equal(order.Items, nil)
	})

	t.Run("expected error from database", func(t *testing.T) {
		is := is.New(t)
		ctrl := gomock.NewController(t)
		mockRepo := bookmock.NewMockRepository(ctrl)
		mockNtfy := bookmock.NewMockNotifier(ctrl)
		mS := book.NewService(mockRepo, mockNtfy, notificationsTimeout)

		newOrderID := uuid.New()
		dbErr := errors.New("fake error from database")
		errRepo := book.ErrResponse{
			Code:    book.ErrResponseFromRespository.Code,
			Message: book.ErrResponseFromRespository.Message + dbErr.Error(),
		}

		mockRepo.EXPECT().ListOrderItems(gomock.Any(), newOrderID).Return(book.Order{}, dbErr)

		order, err := mS.ListOrderItems(ctx, newOrderID)
		is.Equal(err, errRepo)
		is.Equal(order, book.Order{})
		is.Equal(order.Items, nil)
	})

	t.Run("expected context timeout error", func(t *testing.T) {
		is := is.New(t)
		ctrl := gomock.NewController(t)
		mockRepo := bookmock.NewMockRepository(ctrl)
		mockNtfy := bookmock.NewMockNotifier(ctrl)
		mS := book.NewService(mockRepo, mockNtfy, notificationsTimeout)

		newOrderID := uuid.New()

		mockRepo.EXPECT().ListOrderItems(gomock.Any(), newOrderID).Return(book.Order{}, context.DeadlineExceeded)

		order, err := mS.ListOrderItems(ctx, newOrderID)
		is.Equal(err.Error(), "timeout on call to ListOrderItems: "+context.DeadlineExceeded.Error())
		is.Equal(order, book.Order{})
		is.Equal(order.Items, nil)
	})
}

/*
func TestUpdateOrder(t *testing.T) {

	t.Run("updates an order adding a new book", func(t *testing.T) {

		is := is.New(t)
		ctrl := gomock.NewController(t)
		mockRepo := bookmock.NewMockRepository(ctrl)
		mockNtfy := bookmock.NewMockNotifier(ctrl)
		mS := book.NewService(mockRepo, mockNtfy, notificationsTimeout)

		reqOrder := book.UpdateOrderRequest{
			OrderID:   uuid.New(),
			BookID:    uuid.New(),
			BookUnits: 5,
		}

		//UPDATEOrder(book_id, order_id, book_units) order_list
		updatedOrderItemsList, err := mS.UpdateOrder(ctx, reqOrder)
		is.NoErr(err)
		is.Equal(updatedOrderItemsList.Order.OrderID, reqOrder.OrderID)
		//	 MOVE THIS TO EXPECT BLOCK OF MOCKED REPO
		//	is.Equal(updatedOrder.Order_ID, reqOrder.Order_ID)
		//	is.Equal(updatedOrder.Book_ID, reqOrder.Book_ID)
		//	is.Equal(updatedOrder.Book_units, reqOrder.Book_units)
		//	is.True(updatedOrder.UpdatedAt.Compare(updatedOrder.CreatedAt.Round(time.Millisecond)) > 0)

	})
}
*/
