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
	var updatedOrder Order

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
		return Order{}, fmt.Errorf("updating order on db: %w", ErrResponseBookIsArchived)
	}
	balance := *bk.Inventory - updtReq.BookUnitsToAdd
	if balance < 0 {
		return Order{}, fmt.Errorf("updating order on db: %w", ErrResponseInsufficientInventory)
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
