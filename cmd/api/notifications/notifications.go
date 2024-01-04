package notifications

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type Ntfy struct {
	url     string
	enabled bool
	timeout time.Duration
}

func NewNtfy(enableNotifications bool, notificationsTimeout time.Duration) *Ntfy {
	topic := "" //Generate randomic topic
	return &Ntfy{
		url:     fmt.Sprint("https://ntfy.sh/" + topic),
		enabled: enableNotifications,
		timeout: notificationsTimeout,
	}
}

func (ntf *Ntfy) CreatedBookNTF(title string, inventory int) error {
	if ntf.enabled { //Insert context with timeout
		_, err := http.Post(ntf.url, "text/plain", //Create a specific client
			strings.NewReader(fmt.Sprintf("New book created:\nTitle: %s\nInventory: %v", title, inventory)))
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}
