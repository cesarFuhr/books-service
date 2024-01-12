package notifications

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/matryer/is"
)

var ntfy *Ntfy

func TestMain(m *testing.M) {
	notificationsBaseURL := "https://ntfy.sh/test_Ah3mn6oD"
	enableNotifications := true
	ntfy = NewNtfy(enableNotifications, notificationsBaseURL)

	os.Exit(m.Run())
}

func TestBookCreated(t *testing.T) {

	t.Run("notificates the criation of a new book without errors", func(t *testing.T) {
		is := is.New(t)

		title := "book to test ntfy"
		inventory := 35
		notificationsTimeout := 2 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), notificationsTimeout)
		defer cancel()

		//Subscribing to the topic:
		resp, err := ntfy.client.Get(ntfy.baseURL + "_New_book_created/raw")
		is.NoErr(err)
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)

		err = ntfy.BookCreated(ctx, title, inventory)
		is.NoErr(err)

		//Listenning to the topic:
		var message string
		scanner.Scan() //This Scan only gets the opennig of message, that is an empty string!
		scanner.Scan() //This one gets the real message
		message = scanner.Text()

		is.NoErr(scanner.Err())
		is.Equal(message, (fmt.Sprintf("New book created: Title: %s Inventory: %v", title, inventory)))
	})

	t.Run("expected context timeout error", func(t *testing.T) {
		is := is.New(t)

		title := "book to test context timeout"
		inventory := 40
		notificationsTimeout := 2 * time.Millisecond
		ctx, cancel := context.WithTimeout(context.Background(), notificationsTimeout)
		defer cancel()

		err := ntfy.BookCreated(ctx, title, inventory)
		is.True(errors.Is(err, context.DeadlineExceeded))
		is.Equal(err.Error(), `error delivering message (New book created: Title: book to test context timeout Inventory: 40): Post "https://ntfy.sh/test_Ah3mn6oD_New_book_created": context deadline exceeded`)
	})

	t.Run("expected notifications not enabled error", func(t *testing.T) {
		is := is.New(t)
		notificationsBaseURL := "doesn't matter for this test"
		enableNotifications := false
		notNtfy := NewNtfy(enableNotifications, notificationsBaseURL)

		title := "book to test not enabled"
		inventory := 40
		notificationsTimeout := 2 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), notificationsTimeout)
		defer cancel()

		err := notNtfy.BookCreated(ctx, title, inventory)
		is.Equal(err.Error(), "notifications not enabled")
	})
}
