package book

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID           uuid.UUID
	Purchaser_id uuid.UUID
	Order_status string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (s *Service) CreateOrder() Order {
	newOrder := Order{
		ID: uuid.New(),
	}
	return newOrder
}
