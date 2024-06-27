package book

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Order struct {
	Order_ID     uuid.UUID
	Purchaser_ID uuid.UUID
	Order_status string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (s *Service) CreateOrder(ctx context.Context, user_id uuid.UUID) (Order, error) {
	createdAt := time.Now().UTC().Round(time.Millisecond)

	newOrder := Order{
		Order_ID:     uuid.New(),
		Purchaser_ID: user_id,
		Order_status: "accepting_items",
		CreatedAt:    createdAt,
		UpdatedAt:    createdAt,
	}
	return s.repo.CreateOrder(ctx, newOrder)
}

type ItemAtOrder struct {
	Order_ID         uuid.UUID
	Book_ID          uuid.UUID
	Book_units       int
	BookPriceAtOrder float32
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type OrderItemsList struct {
	Order     Order
	ItemsList []ItemAtOrder
}

func (s *Service) ListOrderItems(ctx context.Context, order_id uuid.UUID) (OrderItemsList, error) {
	order, items, err := s.repo.ListOrderItems(ctx, order_id)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return OrderItemsList{Order{}, []ItemAtOrder{}}, fmt.Errorf("timeout on call to ListOrderItems: %w", err)
		}
		errRepo := ErrResponse{
			Code:    ErrResponseFromRespository.Code,
			Message: ErrResponseFromRespository.Message + err.Error(),
		}
		return OrderItemsList{Order{}, []ItemAtOrder{}}, errRepo

	}

	list := OrderItemsList{
		Order:     order,
		ItemsList: items,
	}

	return list, nil
}

type UpdateOrderRequest struct {
	Order_ID   uuid.UUID
	Book_ID    uuid.UUID
	Book_units int
}

/*
func (s *Service) UpdateOrder(ctx context.Context, req UpdateOrderRequest) (Order, error) {
	return updatedOrderItemsList, nil
}
*/
