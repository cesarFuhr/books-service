package notifications_test

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/books-service/cmd/api/book"
	"github.com/books-service/cmd/api/notifications"
	notificationmocks "github.com/books-service/cmd/api/notifications/mocks"
	"github.com/google/uuid"
	"github.com/matryer/is"
	"go.uber.org/mock/gomock"
)

// Integration tests:
func TestIntegrationBookCreated(t *testing.T) {
	notificationsBaseURL := "https://ntfy.sh/test_Ah3mn6oD"
	enableNotifications := true
	testClient := &http.Client{}
	ntfy := notifications.NewNtfy(enableNotifications, notificationsBaseURL, testClient)

	testerBook := book.Book{
		ID:        uuid.New(),
		Name:      "book to test ntfy",
		Price:     toPointer(float32(40.0)),
		Inventory: toPointer(35),
		CreatedAt: time.Now().UTC().Round(time.Millisecond),
		UpdatedAt: time.Now().UTC().Round(time.Millisecond),
	}
	t.Run("notificates the criation of a new book without errors on actual ntfy service", func(t *testing.T) {
		is := is.New(t)

		notificationsTimeout := 2 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), notificationsTimeout)
		defer cancel()

		//Subscribing to the topic:
		resp, err := testClient.Get(notificationsBaseURL + "_New_book_created/raw")
		is.NoErr(err)
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)

		err = ntfy.BookCreated(ctx, testerBook)
		is.NoErr(err)

		//Listenning to the topic:
		var message string
		scanner.Scan() //This Scan only gets the opennig of message, that is an empty string!
		scanner.Scan() //This one gets the real message
		message = scanner.Text()

		is.NoErr(scanner.Err())
		is.Equal(message, (fmt.Sprintf("New book created: ID: %v Title: %s Inventory: %v", testerBook.ID, testerBook.Name, *testerBook.Inventory)))
	})

	t.Run("expected context timeout error", func(t *testing.T) {
		is := is.New(t)

		notificationsTimeout := 2 * time.Millisecond
		ctx, cancel := context.WithTimeout(context.Background(), notificationsTimeout)
		defer cancel()

		err := ntfy.BookCreated(ctx, testerBook)
		is.True(errors.Is(err, context.DeadlineExceeded))
	})
}

// Unit tests:
func TestBookCreated(t *testing.T) {
	notificationsBaseURL := "https://ntfy.sh/test_Ah3mn6oD"
	enableNotifications := true

	testerBook := book.Book{
		ID:        uuid.New(),
		Name:      "book to test ntfy",
		Price:     toPointer(float32(40.0)),
		Inventory: toPointer(35),
		CreatedAt: time.Now().UTC().Round(time.Millisecond),
		UpdatedAt: time.Now().UTC().Round(time.Millisecond),
	}
	t.Run("notificates the criation of a new book without errors on a mocked Client", func(t *testing.T) {
		is := is.New(t)
		ctrl := gomock.NewController(t)
		mockClient := notificationmocks.NewMockDoer(ctrl)
		ntfy := notifications.NewNtfy(enableNotifications, notificationsBaseURL, mockClient)

		ctx := context.Background()

		url := "https://ntfy.sh/test_Ah3mn6oD_New_book_created"
		message := strings.NewReader("New book created:\nID: " + testerBook.ID.String() + "\nTitle: book to test ntfy\nInventory: 35")

		mockClient.EXPECT().Do(gomock.Any()).Do(func(req *http.Request) (*http.Response, error) {
			is.True(req.Method == http.MethodPost)
			is.True(req.URL.String() == url)
			requestedBody, _ := io.ReadAll(req.Body)
			expectedBody, _ := io.ReadAll(message)
			is.Equal(string(requestedBody), string(expectedBody))
			return nil, nil
		})

		err := ntfy.BookCreated(ctx, testerBook)
		is.NoErr(err)
	})

	/*	t.Run("expected context timeout error", func(t *testing.T) {
		is := is.New(t)
		ctrl := gomock.NewController(t)
		mockClient := notificationmocks.NewMockDoer(ctrl)
		ntfy := notifications.NewNtfy(enableNotifications, notificationsBaseURL, mockClient)

		title := "book to test context timeout"
		inventory := 40
		notificationsTimeout := 2 * time.Millisecond
		ctx, cancel := context.WithTimeout(context.Background(), notificationsTimeout)
		defer cancel()

		err := ntfy.BookCreated(ctx, title, inventory)
		is.True(errors.Is(err, context.DeadlineExceeded))
	})*/
}

func toPointer[T any](v T) *T {
	return &v
}
