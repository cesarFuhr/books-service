package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/google/uuid"
)

//==============DATABASE FUNCTIONS:=================

/* Connects to the database trought a connection string and returns a pointer to a valid DB object (*sql.DB). */
func connectDb() *sql.DB {

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
	return db
}

func migrationUp() error {
	driver, err := postgres.WithInstance(dbObject, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migrating up: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("migrating up: %w", err)
	}

	m.Up()
	if err != nil {
		return fmt.Errorf("migrating up: %w", err)
	}
	return nil
}

/* Verifies if there is already a book with the name of "newBook" in the database. If yes, returns it. */
func sameNameOnDB(newBook Book) (unique bool) {
	sqlStatement := `SELECT true FROM bookstable WHERE name=$1;`
	foundRow := dbObject.QueryRow(sqlStatement, newBook.Name)
	var bookAlreadyExists bool
	switch err := foundRow.Scan(&bookAlreadyExists); err {
	case sql.ErrNoRows:
		return true
	case nil:
		return false
	default:
		panic(err)
	}
}

var bookNotFound error

/* Search a book in database based on ID and returns it if succeed. */
func searchById(id uuid.UUID) (Book, error) {
	sqlStatement := `SELECT id, name, price, inventory FROM bookstable WHERE id=$1;`
	foundRow := dbObject.QueryRow(sqlStatement, id)
	var bookToReturn Book
	err := foundRow.Scan(&bookToReturn.ID, &bookToReturn.Name, &bookToReturn.Price, &bookToReturn.Inventory)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			{
				bookNotFound = sql.ErrNoRows //Does it resolve the problem of using SQL on main() side or it's just an useless step?
				return Book{}, fmt.Errorf("searching by ID: %w", bookNotFound)
			}
		default:
			return Book{}, fmt.Errorf("searching by ID: %w", err)
		}
	}

	return bookToReturn, nil
}

/* Stores the book into the database, checks and returns it if succeed. */
func storeOnDB(newBook Book) (fail bool, storedBook Book) {
	sqlStatement := `
	INSERT INTO bookstable (id, name, price, inventory)
	VALUES ($1, $2, $3, $4)
	RETURNING *`
	createdRow := dbObject.QueryRow(sqlStatement, newBook.ID, newBook.Name, *newBook.Price, *newBook.Inventory)
	var bookToReturn Book
	switch err := createdRow.Scan(&bookToReturn.ID, &bookToReturn.Name, &bookToReturn.Price, &bookToReturn.Inventory); err {
	case sql.ErrNoRows:
		return true, Book{}
	case nil:
		return false, bookToReturn
	default:
		panic(err)
	}
}
