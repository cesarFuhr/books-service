package book

import (
	"time"

	bookerrors "github.com/books-service/cmd/api/errors"
	"github.com/google/uuid"
)

type Book struct { //MOVE JSON TAGS TO HTTP PACKAGE!!
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
		return bookerrors.ErrResponseBookEntryBlankFileds
	}
	if bookEntry.Price == nil {
		return bookerrors.ErrResponseBookEntryBlankFileds
	}
	if bookEntry.Inventory == nil {
		return bookerrors.ErrResponseBookEntryBlankFileds
	}

	return nil
}
