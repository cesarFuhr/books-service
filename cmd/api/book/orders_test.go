package book_test

import (
	"context"
	"database/sql"
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

func TestUpdateOrderTX(t *testing.T) {

	t.Run("updates an order adding a new book", func(t *testing.T) {

		is := is.New(t)
		ctrl := gomock.NewController(t)
		mockRepo := bookmock.NewMockRepository(ctrl)
		mockNtfy := bookmock.NewMockNotifier(ctrl)
		mS := book.NewService(mockRepo, mockNtfy, notificationsTimeout)
		mockTxRepo := bookmock.NewMockRepository(ctrl)
		mockTx := bookmock.NewMockTx(ctrl)

		orderToUpdt := book.Order{
			OrderID:     uuid.New(),
			PurchaserID: uuid.New(),
			OrderStatus: "accepting_items",
			CreatedAt:   time.Now().UTC().Round(time.Millisecond),
			UpdatedAt:   time.Now().UTC().Round(time.Millisecond),
			Items:       []book.OrderItem{},
			TotalPrice:  0,
		}
		bkToAdd := book.Book{
			ID: uuid.New(),
			//Name      string
			Price:     toPointer(float32(50.00)),
			Inventory: toPointer(10),
			//CreatedAt time.Time
			//UpdatedAt time.Time
			Archived: false,
		}
		updtReq := book.UpdateOrderRequest{
			OrderID:        orderToUpdt.OrderID,
			BookID:         bkToAdd.ID,
			BookUnitsToAdd: 5,
		}
		createdNow := time.Now().UTC().Round(time.Millisecond)
		newItemAtOrder := book.OrderItem{
			BookID:           bkToAdd.ID,
			BookUnits:        updtReq.BookUnitsToAdd,
			BookPriceAtOrder: bkToAdd.Price,
			CreatedAt:        createdNow,
			UpdatedAt:        createdNow,
		}

		mockRepo.EXPECT().BeginTx(gomock.Any(), nil).Return(mockTxRepo, mockTx, nil)
		mockTxRepo.EXPECT().UpdateOrderRow(gomock.Any(), updtReq.OrderID).Return(nil)
		mockTxRepo.EXPECT().GetBookByID(gomock.Any(), updtReq.BookID).Return(bkToAdd, nil)
		mockTxRepo.EXPECT().UpdateBookAtOrder(gomock.Any(), updtReq).Return(book.OrderItem{}, book.ErrResponseBookNotAtOrder)
		mockTxRepo.EXPECT().AddItemToOrder(gomock.Any(), gomock.Any(), updtReq.OrderID).DoAndReturn(func(ctx context.Context, newItemAtOrder book.OrderItem, orderID uuid.UUID) (book.OrderItem, error) {
			is.Equal(newItemAtOrder.BookID, bkToAdd.ID)
			is.Equal(newItemAtOrder.BookUnits, updtReq.BookUnitsToAdd)
			is.Equal(newItemAtOrder.BookPriceAtOrder, bkToAdd.Price)

			return newItemAtOrder, nil
		})

		*bkToAdd.Inventory = *bkToAdd.Inventory - updtReq.BookUnitsToAdd
		bkToAdd.UpdatedAt = time.Now().UTC().Round(time.Millisecond)
		mockTxRepo.EXPECT().UpdateBook(gomock.Any(), bkToAdd).Return(bkToAdd, nil)
		mockTxRepo.EXPECT().ListOrderItems(gomock.Any(), updtReq.OrderID).DoAndReturn(func(ctx context.Context, order_id uuid.UUID) (book.Order, error) {
			orderToUpdt.UpdatedAt = time.Now().UTC().Round(time.Millisecond).Add(time.Millisecond)
			orderToUpdt.Items = append(orderToUpdt.Items, newItemAtOrder)
			orderToUpdt.TotalPrice = float32(updtReq.BookUnitsToAdd) * *bkToAdd.Price
			return orderToUpdt, nil
		})
		mockTx.EXPECT().Commit().Return(nil)

		mockTx.EXPECT().Rollback().Return(sql.ErrTxDone) //THIS IS ERROR IS NEVER TESTED, ISN'T IT??

		updatedOrder, err := mS.UpdateOrderTx(ctx, updtReq) //MISSING TEST THE TOTALS!!!!
		is.NoErr(err)
		is.Equal(updatedOrder.OrderID, updtReq.OrderID)
		is.True(updatedOrder.UpdatedAt.Compare(updatedOrder.CreatedAt) > 0)
		is.Equal(updatedOrder.Items[0].BookID, updtReq.BookID)
		is.Equal(updatedOrder.Items[0].BookUnits, 5)
		is.Equal(updatedOrder.TotalPrice, float32(250)) //50.00 *  5
		is.True(updatedOrder.Items[0].UpdatedAt.Compare(updatedOrder.Items[0].CreatedAt.Round(time.Millisecond)) == 0)

	})
}
