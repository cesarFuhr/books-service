package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/books-service/cmd/api/book"
	"github.com/books-service/cmd/api/database"
	bookhttp "github.com/books-service/cmd/api/http"
	"github.com/books-service/cmd/api/notifications"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "github.com/lib/pq"
)

func main() {
	err := run()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func run() error {
	//connect to db:
	connStr := os.Getenv("DATABASE_URL")
	dbObject, err := database.ConnectDb(connStr)
	if err != nil {
		return fmt.Errorf("connecting with db: %w", err)
	}

	defer dbObject.Close()

	//apply migrations:
	store := database.NewStore(dbObject)
	path := os.Getenv("DATABASE_MIGRATIONS_PATH")
	err = database.MigrationUp(store, path)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrating: %w", err)
	}

	//get Ntfy notifications config:
	enableNotifications := true
	enableNotificationsStr := os.Getenv("ENABLE_NOTIFICATIONS")
	if enableNotificationsStr != "" {
		enableNotifications, err = strconv.ParseBool(enableNotificationsStr)
		if err != nil {
			return fmt.Errorf("getting notifications enabler flag from env: %w", err)
		}
	}
	notificationsTimeout := 5 * time.Second
	notificationsTimeoutStr := os.Getenv("SERVER_WAITS_NOTIFICATIONS_TIMEOUT") //This ENV must be written with a unit suffix, like seconds
	if notificationsTimeoutStr != "" {
		notificationsTimeout, err = time.ParseDuration(notificationsTimeoutStr)
		if err != nil {
			return fmt.Errorf("getting notifications timeout from env: %w", err)
		}
	}
	notificationsBaseURL := os.Getenv("NOTIFICATIONS_BASE_URL")
	found := strings.HasPrefix(notificationsBaseURL, "https://ntfy.sh/")
	if !found {
		return errors.New("notifications base url must be: https://ntfy.sh/ + some randomic part")
	}

	ntfy := notifications.NewNtfy(enableNotifications, notificationsTimeout, notificationsBaseURL)

	//Init service with its dependencies:
	bookService := book.NewService(store, ntfy)
	bookHandler := bookhttp.NewBookHandler(bookService)

	//create and init http server:
	server := bookhttp.NewServer(bookhttp.ServerConfig{Port: 8080}, bookHandler)

	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("unexpected http server error: %v", err)
		}
		log.Println("stopped serving new requests.")
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc

	timeout := time.Duration(10) * time.Second
	timeoutStr := os.Getenv("SERVICE_SHUTDOWN_TIMEOUT") //This ENV must be written with a unit suffix, like seconds
	if timeoutStr != "" {
		timeout, err = time.ParseDuration(timeoutStr)
		if err != nil {
			return fmt.Errorf("getting shutdown timeout from env: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout))
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("HTTP shutdown error: %w", err)
	}
	log.Println("Graceful shutdown complete.")
	return nil
}
