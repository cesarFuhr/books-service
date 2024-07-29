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
}

func (s *Service) CreateOrder(ctx context.Context, user_id uuid.UUID) (Order, error) {
	createdAt := time.Now().UTC().Round(time.Millisecond)

	newOrder := Order{
		OrderID:     uuid.New(),
		PurchaserID: user_id,
		OrderStatus: "accepting_items",
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}
	return s.repo.CreateOrder(ctx, newOrder)
}

type OrderItem struct {
	OrderID          uuid.UUID
	BookID           uuid.UUID
	BookUnits        int
	BookPriceAtOrder *float32
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type OrderItemsList struct {
	Order     Order
	ItemsList []OrderItem
}

func (s *Service) ListOrderItems(ctx context.Context, order_id uuid.UUID) (OrderItemsList, error) {
	order, items, err := s.repo.ListOrderItems(ctx, order_id)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return OrderItemsList{}, fmt.Errorf("timeout on call to ListOrderItems: %w", err)
		}
		errRepo := ErrResponse{
			Code:    ErrResponseFromRespository.Code,
			Message: ErrResponseFromRespository.Message + err.Error(),
		}
		return OrderItemsList{}, errRepo

	}

	list := OrderItemsList{
		Order:     order,
		ItemsList: items,
	}

	return list, nil
}

type UpdateOrderRequest struct {
	OrderID   uuid.UUID
	BookID    uuid.UUID
	BookUnits int
}

/*
func (s *Service) UpdateOrder(ctx context.Context, req UpdateOrderRequest) (Order, error) {
	return updatedOrderItemsList, nil
}
	BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE;
*/
