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

		mockRepo.EXPECT().ListOrderItems(gomock.Any(), newOrderID).Return(book.Order{}, dbErr)

		order, err := mS.ListOrderItems(ctx, newOrderID)
		is.True(errors.Is(err, dbErr))
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
		is.Equal(err.Error(), "error on call to ListOrderItems: "+context.DeadlineExceeded.Error())
		is.Equal(order, book.Order{})
		is.Equal(order.Items, nil)
	})
}

func TestUpdateOrderTX(t *testing.T) {
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

	t.Run("updates an order adding a new book", func(t *testing.T) {
		is := is.New(t)
		updtReq := book.UpdateOrderRequest{
			OrderID:        orderToUpdt.OrderID,
			BookID:         bkToAdd.ID,
			BookUnitsToAdd: 5,
		}
		createdNow := time.Now().UTC().Round(time.Millisecond)
		newOrderItem := book.OrderItem{
			BookID:           bkToAdd.ID,
			BookUnits:        updtReq.BookUnitsToAdd,
			BookPriceAtOrder: bkToAdd.Price,
		}

		mockRepo.EXPECT().BeginTx(gomock.Any(), nil).Return(mockTxRepo, mockTx, nil)
		mockTxRepo.EXPECT().UpdateOrderRow(gomock.Any(), updtReq.OrderID).DoAndReturn(func(context.Context, uuid.UUID) error {
			orderToUpdt.UpdatedAt = time.Now().UTC().Round(time.Millisecond).Add(time.Millisecond)
			return nil
		})
		mockTxRepo.EXPECT().GetBookByID(gomock.Any(), updtReq.BookID).Return(bkToAdd, nil)
		mockTxRepo.EXPECT().GetOrderItem(gomock.Any(), updtReq.OrderID, updtReq.BookID).Return(book.OrderItem{}, book.ErrResponseBookNotAtOrder)
		mockTxRepo.EXPECT().UpsertOrderItem(gomock.Any(), updtReq.OrderID, newOrderItem).DoAndReturn(func(context.Context, uuid.UUID, book.OrderItem) (book.OrderItem, error) {
			newOrderItem = book.OrderItem{
				BookID:           bkToAdd.ID,
				BookUnits:        updtReq.BookUnitsToAdd,
				BookPriceAtOrder: bkToAdd.Price,
				CreatedAt:        createdNow,
				UpdatedAt:        createdNow,
			}
			return newOrderItem, nil
		})

		mockTxRepo.EXPECT().UpdateBook(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, bookEntry book.Book) (book.Book, error) {
			is.Equal(bookEntry.ID, bkToAdd.ID)
			is.Equal(bookEntry.Price, bkToAdd.Price)
			is.Equal(bookEntry.Inventory, bkToAdd.Inventory)
			bkToAdd.UpdatedAt = time.Now().UTC().Round(time.Millisecond)
			return bkToAdd, nil
		})
		mockTxRepo.EXPECT().ListOrderItems(gomock.Any(), updtReq.OrderID).DoAndReturn(func(ctx context.Context, order_id uuid.UUID) (book.Order, error) {
			orderToUpdt.Items = append(orderToUpdt.Items, newOrderItem)
			orderToUpdt.TotalPrice = float32(updtReq.BookUnitsToAdd) * *bkToAdd.Price
			return orderToUpdt, nil
		})
		mockTx.EXPECT().Commit().Return(nil)

		mockTx.EXPECT().Rollback().Return(sql.ErrTxDone) //THIS ERROR IS NEVER TESTED, ISN'T IT??

		updatedOrder, err := mS.UpdateOrderTx(ctx, updtReq)
		is.NoErr(err)
		is.Equal(updatedOrder.OrderID, updtReq.OrderID)
		is.True(updatedOrder.UpdatedAt.Compare(updatedOrder.CreatedAt) > 0)
		is.Equal(updatedOrder.Items[0].BookID, updtReq.BookID)
		is.Equal(updatedOrder.Items[0].BookUnits, 5)
		is.Equal(updatedOrder.TotalPrice, float32(250)) //50.00 *  5
		is.True(updatedOrder.Items[0].UpdatedAt.Compare(updatedOrder.Items[0].CreatedAt.Round(time.Millisecond)) == 0)
		is.Equal(*bkToAdd.Inventory, 5) //10 - 5 = 5
	})

	t.Run("adds units to a book that is already at order", func(t *testing.T) {
		is := is.New(t)
		updtReq := book.UpdateOrderRequest{
			OrderID:        orderToUpdt.OrderID,
			BookID:         bkToAdd.ID,
			BookUnitsToAdd: 5,
		}

		mockRepo.EXPECT().BeginTx(gomock.Any(), nil).Return(mockTxRepo, mockTx, nil)
		mockTxRepo.EXPECT().UpdateOrderRow(gomock.Any(), updtReq.OrderID).DoAndReturn(func(context.Context, uuid.UUID) error {
			orderToUpdt.UpdatedAt = time.Now().UTC().Round(time.Millisecond).Add(time.Millisecond)
			return nil
		})
		mockTxRepo.EXPECT().GetBookByID(gomock.Any(), updtReq.BookID).Return(bkToAdd, nil)
		mockTxRepo.EXPECT().GetOrderItem(gomock.Any(), updtReq.OrderID, updtReq.BookID).Return(orderToUpdt.Items[0], nil)

		orderItemToUpdate := book.OrderItem{
			BookID:           orderToUpdt.Items[0].BookID,
			BookUnits:        orderToUpdt.Items[0].BookUnits + updtReq.BookUnitsToAdd, //Must result 10
			BookPriceAtOrder: orderToUpdt.Items[0].BookPriceAtOrder,
		}

		mockTxRepo.EXPECT().UpsertOrderItem(gomock.Any(), updtReq.OrderID, gomock.Any()).DoAndReturn(func(ctx context.Context, id uuid.UUID, oItem book.OrderItem) (book.OrderItem, error) {
			is.Equal(orderItemToUpdate.BookID, orderToUpdt.Items[0].BookID)
			is.Equal(orderItemToUpdate.BookUnits, orderToUpdt.Items[0].BookUnits+updtReq.BookUnitsToAdd)
			is.Equal(orderItemToUpdate.BookPriceAtOrder, orderToUpdt.Items[0].BookPriceAtOrder)

			orderToUpdt.Items[0].BookUnits = orderItemToUpdate.BookUnits
			orderToUpdt.Items[0].UpdatedAt = time.Now().UTC().Round(time.Millisecond).Add(time.Millisecond)
			return orderToUpdt.Items[0], nil
		})

		mockTxRepo.EXPECT().UpdateBook(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, bookEntry book.Book) (book.Book, error) {
			is.Equal(bookEntry.ID, bkToAdd.ID)
			is.Equal(bookEntry.Price, bkToAdd.Price)
			is.Equal(bookEntry.Inventory, bkToAdd.Inventory)
			bkToAdd.UpdatedAt = time.Now().UTC().Round(time.Millisecond)
			return bkToAdd, nil
		})
		mockTxRepo.EXPECT().ListOrderItems(gomock.Any(), updtReq.OrderID).DoAndReturn(func(ctx context.Context, order_id uuid.UUID) (book.Order, error) {
			orderToUpdt.TotalPrice = float32(orderToUpdt.Items[0].BookUnits) * *orderToUpdt.Items[0].BookPriceAtOrder
			return orderToUpdt, nil
		})

		mockTx.EXPECT().Commit().Return(nil)

		mockTx.EXPECT().Rollback().Return(sql.ErrTxDone)

		updatedOrder, err := mS.UpdateOrderTx(ctx, updtReq)
		is.NoErr(err)
		is.Equal(updatedOrder.OrderID, updtReq.OrderID)
		is.True(updatedOrder.UpdatedAt.Compare(updatedOrder.CreatedAt) > 0)
		is.Equal(updatedOrder.Items[0].BookID, updtReq.BookID)
		is.Equal(updatedOrder.Items[0].BookUnits, 10)
		is.Equal(updatedOrder.TotalPrice, float32(500)) //50.00 *  10
		is.True(updatedOrder.Items[0].UpdatedAt.Compare(updatedOrder.Items[0].CreatedAt.Round(time.Millisecond)) > 0)
		is.Equal(*bkToAdd.Inventory, 0)

	})
	t.Run("removes a book from the order", func(t *testing.T) {
		is := is.New(t)

		updtReq := book.UpdateOrderRequest{
			OrderID:        orderToUpdt.OrderID,
			BookID:         bkToAdd.ID,
			BookUnitsToAdd: -10,
		}

		mockRepo.EXPECT().BeginTx(gomock.Any(), nil).Return(mockTxRepo, mockTx, nil)
		mockTxRepo.EXPECT().UpdateOrderRow(gomock.Any(), updtReq.OrderID).DoAndReturn(func(context.Context, uuid.UUID) error {
			orderToUpdt.UpdatedAt = time.Now().UTC().Round(time.Millisecond).Add(time.Millisecond)
			return nil
		})
		mockTxRepo.EXPECT().GetBookByID(gomock.Any(), updtReq.BookID).Return(bkToAdd, nil)
		mockTxRepo.EXPECT().GetOrderItem(gomock.Any(), updtReq.OrderID, updtReq.BookID).Return(orderToUpdt.Items[0], nil)
		mockTxRepo.EXPECT().DeleteOrderItem(gomock.Any(), updtReq.OrderID, updtReq.BookID).Return(nil)

		mockTxRepo.EXPECT().UpdateBook(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, bookEntry book.Book) (book.Book, error) {
			is.Equal(bookEntry.ID, bkToAdd.ID)
			is.Equal(bookEntry.Price, bkToAdd.Price)
			is.Equal(bookEntry.Inventory, bkToAdd.Inventory)
			bkToAdd.UpdatedAt = time.Now().UTC().Round(time.Millisecond)
			return bkToAdd, nil
		})
		mockTxRepo.EXPECT().ListOrderItems(gomock.Any(), updtReq.OrderID).DoAndReturn(func(ctx context.Context, order_id uuid.UUID) (book.Order, error) {
			orderToUpdt.Items = []book.OrderItem{}
			orderToUpdt.TotalPrice = float32(0)
			return orderToUpdt, nil
		})

		mockTx.EXPECT().Commit().Return(nil)

		mockTx.EXPECT().Rollback().Return(sql.ErrTxDone)

		updatedOrder, err := mS.UpdateOrderTx(ctx, updtReq)
		is.NoErr(err)
		is.Equal(updatedOrder.OrderID, updtReq.OrderID)
		is.True(updatedOrder.UpdatedAt.Compare(updatedOrder.CreatedAt) > 0)
		is.Equal(updatedOrder.Items, []book.OrderItem{})
		is.Equal(updatedOrder.TotalPrice, float32(0))
		is.Equal(*bkToAdd.Inventory, 10)
	})
}
