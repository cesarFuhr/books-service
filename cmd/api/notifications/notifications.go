package notifications

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type Ntfy struct {
	baseURL string
	enabled bool
	client  *http.Client
}

func NewNtfy(enableNotifications bool, notificationsBaseURL string) *Ntfy {
	return &Ntfy{
		baseURL: notificationsBaseURL,
		enabled: enableNotifications,
		client:  &http.Client{},
	}
}

func (ntf *Ntfy) BookCreated(ctx context.Context, title string, inventory int) error {
	if ntf.enabled {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, ntf.baseURL+"_New_book_created", strings.NewReader(fmt.Sprintf("New book created:\nTitle: %s\nInventory: %v", title, inventory))) //Ntfy SEEMS NOT TO ACEPT SLASHS OR DOTS AT TOPIC
		if err != nil {
			return fmt.Errorf("error delivering message ("+fmt.Sprintf("New book created: Title: %s Inventory: %v", title, inventory)+"): %w", err)
		}
		_, err = ntf.client.Do(req)
		if err != nil {
			return fmt.Errorf("error delivering message ("+fmt.Sprintf("New book created: Title: %s Inventory: %v", title, inventory)+"): %w", err)
		}
		return nil
	}
	return errors.New("notifications not enabled")
}
