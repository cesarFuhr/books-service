package notifications

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type Ntfy struct {
	url     string
	enabled bool
	timeout time.Duration
	client  *http.Client
}

func NewNtfy(enableNotifications bool, notificationsTimeout time.Duration) *Ntfy {
	randomTopic := randomString(8)                             //Generate randomic part of topic
	log.Printf("random part of ntfy topic: %s\n", randomTopic) //SHOULD WE EXPORT THIS VALUE TO AN ENV VARIABLE SO WE CAN PASTE IT TO NTFY.APP?
	var ntfClient *http.Client = &http.Client{}                //SHOULD WE ADD SOME SPECIFIC CONFIGURATION TO THIS CLIENT? By now, its just a DefaultClient
	return &Ntfy{
		url:     fmt.Sprint("https://ntfy.sh/" + randomTopic),
		enabled: enableNotifications,
		timeout: notificationsTimeout,
		client:  ntfClient,
	}
}

func (ntf *Ntfy) BookCreated(title string, inventory int) error {
	if ntf.enabled {
		ctx, cancel := context.WithTimeout(context.Background(), ntf.timeout)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, ntf.url+"/New_book_created", strings.NewReader(fmt.Sprintf("New book created:\nTitle: %s\nInventory: %v", title, inventory)))
		if err != nil {
			return fmt.Errorf("error delivering message ("+fmt.Sprintf("New book created:\nTitle: %s\nInventory: %v", title, inventory)+") to topic ("+ntf.url+"/New_book_created): %w", err)
		}
		_, err = ntf.client.Do(req)
		if err != nil {
			return fmt.Errorf("error delivering message ("+fmt.Sprintf("New book created:\nTitle: %s\nInventory: %v", title, inventory)+") to topic ("+ntf.url+"/New_book_created): %w", err)
		}
	}
	return errors.New("notifications not enabled")
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
