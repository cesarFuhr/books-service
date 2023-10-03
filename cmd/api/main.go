package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/books-service/cmd/api/database"
	"github.com/books-service/cmd/api/pkghttp"

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
	connStr := os.Getenv("DATABASE_URL") //I'M STILL CONFUSED ABOUT THAT. IN DATABASE_TEST WE CHANGED THE ENV BY A LITERAL
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

	//create and init http server:
	server := pkghttp.NewServer(pkghttp.ServerConfig{Port: 8080})

	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("unexpected http server error: %w", err)
	}
	return nil
}
