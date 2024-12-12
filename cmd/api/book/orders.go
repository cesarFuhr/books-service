package book

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Order struct {
	OrderID     uuid.UUID
	PurchaserID uuid.UUID
	OrderStatus string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Items       []OrderItem
	TotalPrice  float32
}

func (s *Service) CreateOrder(ctx context.Context, user_id uuid.UUID) (Order, error) {

	createdAt := time.Now().UTC().Round(time.Millisecond)

	newOrder := Order{
		OrderID:     uuid.New(),
		PurchaserID: user_id,
		OrderStatus: "accepting_items",
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
		Items:       []OrderItem{},
	}
	return s.repo.CreateOrder(ctx, newOrder)
}

type OrderItem struct {
	BookID           uuid.UUID
	BookUnits        int
	BookPriceAtOrder *float32
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (s *Service) ListOrderItems(ctx context.Context, order_id uuid.UUID) (Order, error) {

	order, err := s.repo.ListOrderItems(ctx, order_id)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return Order{}, fmt.Errorf("timeout on call to ListOrderItems: %w", err)
		}
		errRepo := ErrResponse{
			Code:    ErrResponseFromRespository.Code,
			Message: ErrResponseFromRespository.Message + err.Error(),
		}
		return Order{}, errRepo

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
		if errors.Is(err, context.DeadlineExceeded) {
			return Order{}, fmt.Errorf("timeout on call to UpdateOrderTx: %w ", err)
		}
		errRepo := ErrResponse{
			Code:    ErrResponseFromRespository.Code,
			Message: ErrResponseFromRespository.Message + err.Error(),
		}
		return Order{}, errRepo
	}

	defer func() {
		tx.Rollback()
	}()

	err = txRepo.UpdateOrderRow(ctx, updtReq.OrderID) //changes field 'updated_at' and checks if the order is 'accepting_items'
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return Order{}, fmt.Errorf("timeout on call to UpdateOrderTx: %w ", err)
		} else if errors.Is(err, ErrResponseOrderNotFound) {
			return Order{}, ErrResponseOrderNotFound
		} else if errors.Is(err, ErrResponseOrderNotAcceptingItems) {
			return Order{}, ErrResponseOrderNotAcceptingItems
		}
		errRepo := ErrResponse{
			Code:    ErrResponseFromRespository.Code,
			Message: ErrResponseFromRespository.Message + err.Error(),
		}
		return Order{}, errRepo
	}

	//Testing if there are sufficient inventory of the book asked, and if is not archived:
	bk, err := txRepo.GetBookByID(ctx, updtReq.BookID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return Order{}, fmt.Errorf("timeout on call to UpdateOrderTx: %w ", err)
		}
		errRepo := ErrResponse{
			Code:    ErrResponseFromRespository.Code,
			Message: ErrResponseFromRespository.Message + err.Error(),
		}
		return Order{}, errRepo
	}
	if bk.Archived {
		return Order{}, ErrResponseBookIsArchived
	}
	balance := *bk.Inventory - updtReq.BookUnitsToAdd
	if balance < 0 {
		return Order{}, ErrResponseInsufficientInventory
	}

	//Testing if the book is already at the order and, if it is, updating it:
	bookAtOrder, err := txRepo.UpdateBookAtOrder(ctx, updtReq)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return Order{}, fmt.Errorf("timeout on call to UpdateOrderTx: %w ", err)
		} else {
			if errors.Is(err, ErrResponseBookNotAtOrder) {
				if updtReq.BookUnitsToAdd > 0 { //Considering that the book is not at the order yet, this call to UpdateOrderTX only makes sense if BookUnitsToAdd is positive.
					//Adding a new book to the order:
					createdNow := time.Now().UTC().Round(time.Millisecond)
					newItemAtOrder := OrderItem{
						BookID:           updtReq.BookID,
						BookUnits:        updtReq.BookUnitsToAdd,
						BookPriceAtOrder: bk.Price,
						CreatedAt:        createdNow,
						UpdatedAt:        createdNow,
					}

					bookAtOrder, err = txRepo.AddItemToOrder(ctx, newItemAtOrder, updtReq.OrderID)
					if err != nil {
						if errors.Is(err, context.DeadlineExceeded) {
							return Order{}, fmt.Errorf("timeout on call to UpdateOrderTx: %w ", err)
						}
						errRepo := ErrResponse{
							Code:    ErrResponseFromRespository.Code,
							Message: ErrResponseFromRespository.Message + err.Error(),
						}
						return Order{}, errRepo
					}
				} else {
					return Order{}, ErrResponseUpdateRequestNotPositive
				}
			} else {
				errRepo := ErrResponse{
					Code:    ErrResponseFromRespository.Code,
					Message: ErrResponseFromRespository.Message + err.Error(),
				}
				return Order{}, errRepo
			}
		}
	} else { //Case the book is already at the order, and book_units becomes zero from update, the book is excluded from the order. Even so, it must be updated at bookstable.
		if bookAtOrder.BookUnits == 0 {
			err = txRepo.DeleteBookAtOrder(ctx, updtReq)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					return Order{}, fmt.Errorf("timeout on call to UpdateOrderTx: %w ", err)
				}
				errRepo := ErrResponse{
					Code:    ErrResponseFromRespository.Code,
					Message: ErrResponseFromRespository.Message + err.Error(),
				}
				return Order{}, errRepo
			}
		}
	}

	//Updating book inventory acordingly at bookstable:
	*bk.Inventory = balance
	bk.UpdatedAt = time.Now().UTC().Round(time.Millisecond)
	_, err = txRepo.UpdateBook(ctx, bk)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return Order{}, fmt.Errorf("timeout on call to UpdateOrderTx: %w ", err)
		}
		errRepo := ErrResponse{
			Code:    ErrResponseFromRespository.Code,
			Message: ErrResponseFromRespository.Message + err.Error(),
		}
		return Order{}, errRepo
	}

	updatedOrder, err := txRepo.ListOrderItems(ctx, updtReq.OrderID) //AND THE TOTALS?????
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return Order{}, fmt.Errorf("timeout on call to UpdateOrderTx: %w ", err)
		}
		errRepo := ErrResponse{
			Code:    ErrResponseFromRespository.Code,
			Message: ErrResponseFromRespository.Message + err.Error(),
		}
		return Order{}, errRepo
	}

	err = tx.Commit()
	if err != nil {
		errRepo := ErrResponse{
			Code:    ErrResponseFromRespository.Code,
			Message: ErrResponseFromRespository.Message + err.Error(),
		}
		return Order{}, errRepo
	}

	return updatedOrder, nil

}
