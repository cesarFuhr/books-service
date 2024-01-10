package notifications

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Ntfy struct {
	baseURL string
	enabled bool
	timeout time.Duration
	client  *http.Client
}

func NewNtfy(enableNotifications bool, notificationsTimeout time.Duration, notificationsBaseURL string) *Ntfy {
	return &Ntfy{
		baseURL: notificationsBaseURL,
		enabled: enableNotifications,
		timeout: notificationsTimeout,
		client:  &http.Client{},
	}
}

func (ntf *Ntfy) BookCreated(title string, inventory int) error {
	if ntf.enabled {
		ctx, cancel := context.WithTimeout(context.Background(), ntf.timeout)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, ntf.baseURL+"/New_book_created", strings.NewReader(fmt.Sprintf("New book created:\nTitle: %s\nInventory: %v", title, inventory)))
		if err != nil {
			return fmt.Errorf("error delivering message ("+fmt.Sprintf("New book created:\nTitle: %s\nInventory: %v", title, inventory)+") to topic ("+ntf.baseURL+"/New_book_created): %w", err)
		}
		_, err = ntf.client.Do(req)
		if err != nil {
			return fmt.Errorf("error delivering message ("+fmt.Sprintf("New book created:\nTitle: %s\nInventory: %v", title, inventory)+") to topic ("+ntf.baseURL+"/New_book_created): %w", err)
		}
	}
	return errors.New("notifications not enabled")
}
