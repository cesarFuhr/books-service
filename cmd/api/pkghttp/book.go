package pkghttp

import (
	"time"

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
func filledFields(bookEntry Book) error {
	if bookEntry.Name == "" {
		return errResponseBookEntryBlankFileds
	}
	if bookEntry.Price == nil {
		return errResponseBookEntryBlankFileds
	}
	if bookEntry.Inventory == nil {
		return errResponseBookEntryBlankFileds
	}

	return nil
}
