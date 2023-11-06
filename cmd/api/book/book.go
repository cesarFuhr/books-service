package book

import (
	"time"

	"github.com/google/uuid"
)

const PriceMax = 9999.99 //max value to field price on db, set to: numeric(6,2)

type Book struct { //IS IT POSSIBLE TO MOVE JSON TAGS TO HTTP PACKAGE??
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Price     *float32  `json:"price"`
	Inventory *int      `json:"inventory"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	Archived  bool      `json:"archived"`
}

/* Verifies if all entry fields are filled and returns a warning message if so. */
func FilledFields(bookEntry EntryBookRequest) error {
	if bookEntry.Name == "" {
		return ErrResponseBookEntryBlankFileds
	}
	if bookEntry.Price == nil {
		return ErrResponseBookEntryBlankFileds
	}
	if bookEntry.Inventory == nil {
		return ErrResponseBookEntryBlankFileds
	}

	return nil
}
