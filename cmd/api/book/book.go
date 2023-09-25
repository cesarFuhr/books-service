package book

import (
	"time"

	"github.com/books-service/cmd/api/pkgerrors"
	"github.com/google/uuid"
)

type Book struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Price     *float32  `json:"price"`
	Inventory *int      `json:"inventory"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	Archived  bool      `json:"archived"`
}

/* Verifies if all entry fields are filled and returns a warning message if so. */
func FilledFields(bookEntry Book) error {
	if bookEntry.Name == "" {
		return pkgerrors.ErrResponseBookEntryBlankFileds
	}
	if bookEntry.Price == nil {
		return pkgerrors.ErrResponseBookEntryBlankFileds
	}
	if bookEntry.Inventory == nil {
		return pkgerrors.ErrResponseBookEntryBlankFileds
	}

	return nil
}
