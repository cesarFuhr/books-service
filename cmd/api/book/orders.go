package book

import (
	"context"
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
	return newOrder, nil
}
