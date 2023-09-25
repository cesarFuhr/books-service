package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "github.com/lib/pq"
)

var dbObjectGlobal *sql.DB

func main() {
	//connect to db:
	dbObject, err := connectDb()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	dbObjectGlobal = dbObject
	defer dbObjectGlobal.Close()

	//apply migrations:
	err = migrationUp()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Println(err)
		os.Exit(1)
	}

	//start http server:
	http.HandleFunc("/ping", ping)
	http.HandleFunc("/books", books)
	http.HandleFunc("/books/", bookById)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
