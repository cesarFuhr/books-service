package book

import (
	"time"

	"github.com/google/uuid"
)

const PriceMax = 9999.99 //max value to field price on db, set to: numeric(6,2)

type Book struct {
	ID        uuid.UUID
	Name      string
	Price     *float32
	Inventory *int
	CreatedAt time.Time
	UpdatedAt time.Time
	Archived  bool
}
