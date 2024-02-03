package notifications

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/books-service/cmd/api/book"
)

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type Ntfy struct {
	baseURL string
	enabled bool
	client  Doer
}

func NewNtfy(enableNotifications bool, notificationsBaseURL string, client Doer) *Ntfy {
	return &Ntfy{
		baseURL: notificationsBaseURL,
		enabled: enableNotifications,
		client:  client,
	}
}

func (ntf *Ntfy) BookCreated(ctx context.Context, createdBook book.Book) error {
	if !ntf.enabled {
		return nil
	}

	url := ntf.baseURL + "_New_book_created"
	message := strings.NewReader(fmt.Sprintf("New book created:\nID: %v\nTitle: %s\nInventory: %v", createdBook.ID, createdBook.Name, *createdBook.Inventory)) //Ntfy SEEMS NOT TO ACEPT SLASHS OR DOTS AT TOPIC

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, message)
	if err != nil {
		return fmt.Errorf("error delivering message to ntfy (book ID: %v): %w", createdBook.ID, err)
	}

	resp, err := ntf.client.Do(req)
	if err != nil {
		return fmt.Errorf("error delivering message to ntfy (book ID: %v): %w", createdBook.ID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		book.ErrStatusNotOK = book.ErrStatus{
			StatusCode: resp.StatusCode,
			Message:    "ntfy wrong response - want: 200 OK, got: " + resp.Status,
		}
		return book.ErrStatusNotOK
	}

	return nil
}
