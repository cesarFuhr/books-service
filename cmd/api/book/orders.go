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

/* WRITE THIS FUNCTION LATER...
func (s *Service) UpdateOrder(ctx context.Context, req UpdateOrderRequest) (OrderItemsList, error) {
	var updatedOrderItemsList OrderItemsList

	return updatedOrderItemsList, nil
}
*/
