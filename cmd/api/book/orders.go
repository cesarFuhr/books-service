package book

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

type Order struct {
	OrderID     uuid.UUID
	PurchaserID uuid.UUID
	OrderStatus string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	TotalPrice  float32
	Items       []OrderItem
}

func (s *Service) CreateOrder(ctx context.Context, user_id uuid.UUID) (Order, error) {

	createdAt := time.Now().UTC().Round(time.Millisecond)

	newOrder := Order{
		OrderID:     uuid.New(),
		PurchaserID: user_id,
		OrderStatus: "accepting_items",
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
		TotalPrice:  0,
		Items:       []OrderItem{},
	}
	return s.repo.CreateOrder(ctx, newOrder)
}

type OrderItem struct {
	BookID           uuid.UUID
	BookName         string
	BookUnits        int
	BookPriceAtOrder *float32
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (s *Service) ListOrderItems(ctx context.Context, order_id uuid.UUID) (Order, error) {

	order, err := s.repo.ListOrderItems(ctx, order_id)
	if err != nil {
		return Order{}, fmt.Errorf("error on call to ListOrderItems: %w", err)
	}

	return order, nil
}

type UpdateOrderRequest struct {
	OrderID        uuid.UUID
	BookID         uuid.UUID
	BookUnitsToAdd int
}

/* Updates an order stored in database through a transaction, adding or removing items(books) from it. */
func (s *Service) UpdateOrderTx(ctx context.Context, updtReq UpdateOrderRequest) (Order, error) {
	txRepo, tx, err := s.repo.BeginTx(ctx, nil)
	if err != nil {
		return Order{}, fmt.Errorf("error on call to BeginTx: %w ", err)
	}

	defer func() {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil && rollbackErr != sql.ErrTxDone {
			log.Println(rollbackErr)
		}
	}()

	err = txRepo.UpdateOrderRow(ctx, updtReq.OrderID) //changes field 'updated_at' and checks if the order is 'accepting_items'
	if err != nil {
		return Order{}, fmt.Errorf("error on call to UpdateOrderRow: %w ", err)
	}

	//Testing if there are sufficient inventory of the book asked, and if is not archived:
	bk, err := txRepo.GetBookByID(ctx, updtReq.BookID)
	if err != nil {
		if errors.Is(err, ErrResponseBookNotFound) {
			return Order{}, ErrResponseBookNotFound
		}
		return Order{}, fmt.Errorf("error on call to GetBookByID: %w ", err)
	}
	if bk.Archived {
		return Order{}, ErrResponseBookIsArchived
	}
	if *bk.Inventory-updtReq.BookUnitsToAdd < 0 {
		return Order{}, ErrResponseInsufficientInventory
	}

	//Testing if the book is already at the order and, if it is, getting it:
	bookAtOrder, err := txRepo.GetOrderItem(ctx, updtReq.OrderID, updtReq.BookID)
	if err != nil {
		if errors.Is(err, ErrResponseBookNotAtOrder) && updtReq.BookUnitsToAdd <= 0 { //But, if the book is not at order, a request attempting to decrease its units value can mean an error from client, so an error is returned.
			return Order{}, ErrResponseBookNotAtOrder
		}
		if !errors.Is(err, ErrResponseBookNotAtOrder) {
			return Order{}, fmt.Errorf("error on call to GetOrderItem: %w ", err)
		}
	}

	//Calculating changes to order item:
	updtBookUnits := bookAtOrder.BookUnits + updtReq.BookUnitsToAdd

	if updtBookUnits > 0 { //This way means that the book is being added to the order or, after any changes, some units of it remain there.

		//Adding or updating the book at the order:
		bookAtOrder.BookID = updtReq.BookID
		bookAtOrder.BookName = bk.Name
		bookAtOrder.BookUnits = updtBookUnits
		if bookAtOrder.BookPriceAtOrder == nil { //If the book is already at the order, this price must be maintenned.
			bookAtOrder.BookPriceAtOrder = bk.Price
		}
		//Created_at and Updated_at fields will be set properly at database layer

		bookAtOrder, err = txRepo.UpsertOrderItem(ctx, updtReq.OrderID, bookAtOrder)
		if err != nil {
			return Order{}, fmt.Errorf("error on call to UpsertOrderItem: %w ", err)
		}

		//Updating book inventory acordingly at bookstable:
		*bk.Inventory = *bk.Inventory - updtReq.BookUnitsToAdd

	} else { //Case the book is already at the order, and book_units becomes zero from update, the book is excluded from the order. Even so, it must be updated at bookstable.

		err = txRepo.DeleteOrderItem(ctx, updtReq.OrderID, updtReq.BookID)

		if err != nil {
			return Order{}, fmt.Errorf("error on call to DeleteOrderItem: %w ", err)
		}

		//Updating book inventory acordingly at bookstable:
		*bk.Inventory += bookAtOrder.BookUnits
	}

	bk.UpdatedAt = time.Now().UTC().Round(time.Millisecond)
	_, err = txRepo.UpdateBook(ctx, bk)
	if err != nil {
		return Order{}, fmt.Errorf("error on call to UpdateBook: %w ", err)
	}

	err = tx.Commit()
	if err != nil {
		return Order{}, fmt.Errorf("error on call to Commit: %w ", err)
	}

	updatedOrder, err := s.repo.ListOrderItems(ctx, updtReq.OrderID)
	if err != nil {
		return Order{}, fmt.Errorf("error on call to ListOrderItems: %w ", err)
	}

	return updatedOrder, nil
}
