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

	bookService := book.NewService(store)
	bookHandler := bookhttp.NewBookHandler(bookService)

	//create and init http server:
	server := bookhttp.NewServer(bookhttp.ServerConfig{Port: 8080}, bookHandler)

	go func() (err error) {
		err = server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("unexpected http server error: %w", err)
		}
		return nil
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc

	ctx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second) //HOW MUCH TIME MAKE SENSE FOR US??
	defer shutdownRelease()
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("HTTP shutdown error: %w", err)
	}
	log.Println("Graceful shutdown complete.")
	return err
}
