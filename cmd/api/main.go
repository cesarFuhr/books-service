package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/books-service/cmd/api/book"
	"github.com/books-service/cmd/api/database"
	bookhttp "github.com/books-service/cmd/api/http"
	"github.com/books-service/cmd/api/notifications"

	"github.com/golang-migrate/migrate/v4"
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

	//get request timeout from environment:
	reqTimeout := time.Duration(5) * time.Second
	reqTimeoutStr := os.Getenv("HTTP_REQUEST_TIMEOUT") //This ENV must be written with a unit suffix, like seconds
	if reqTimeoutStr != "" {
		reqTimeout, err = time.ParseDuration(reqTimeoutStr)
		if err != nil {
			return fmt.Errorf("getting request timeout from env: %w", err)
		}
	}

	//get Ntfy notifications config:
	enableNotifications := true             //GET FROM ENV!!!!!!!!!
	notificationsTimeout := 2 * time.Second //GET FROM ENV!!!!!!!!
	ntfy := notifications.NewNtfy(enableNotifications, notificationsTimeout)

	//Init service with its dependencies:
	bookService := book.NewService(store, ntfy)
	bookHandler := bookhttp.NewBookHandler(bookService, reqTimeout)

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
