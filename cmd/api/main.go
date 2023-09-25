package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "github.com/lib/pq"
)

var dbObjectGlobal *sql.DB

func main() {
	err := run()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func run() error {
	//connect to db:
	dbObject, err := connectDb()
	if err != nil {
		return fmt.Errorf("connecting with db: %w", err)
	}

	dbObjectGlobal = dbObject
	defer dbObjectGlobal.Close()

	//apply migrations:
	err = migrationUp()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrating: %w", err)
	}

	//create and init http server:
	server := newServer(serverConfig{Port: 8080})

	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("unexpected http server error: %w", err)
	}
	return nil
}
