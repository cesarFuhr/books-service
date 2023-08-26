package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/google/uuid"
)

//==============DATABASE FUNCTIONS:=================

/* Connects to the database trought a connection string and returns a pointer to a valid DB object (*sql.DB). */
func connectDb() (*sql.DB, error) {

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return db, fmt.Errorf("connecting to db: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return db, fmt.Errorf("connecting to db: %w", err)
	}

	fmt.Println("Successfully connected!")
	return db, nil
}

func migrationUp() error {
	driver, err := postgres.WithInstance(dbObjectGlobal, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migrating up: %w", err)
	}

	path := os.Getenv("DATABASE_MIGRATIONS_PATH")
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", path),
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
func sameNameOnDB(newBook Book) (unique bool, unexpected error) {
	sqlStatement := `SELECT true FROM bookstable WHERE name=$1;`
	foundRow := dbObjectGlobal.QueryRow(sqlStatement, newBook.Name)
	var bookAlreadyExists bool
	err := foundRow.Scan(&bookAlreadyExists)
	switch err {
	case sql.ErrNoRows:
		return true, nil
	case nil:
		return false, nil
	default:
		return false, fmt.Errorf("verifying same name on db: %w", err)
	}
}

var errBookNotFound = errors.New("book not found")

/* Searches a book in database based on ID and returns it if succeed. */
func searchById(id uuid.UUID) (Book, error) {
	sqlStatement := `SELECT id, name, price, inventory, created_at, updated_at FROM bookstable WHERE id=$1;`
	foundRow := dbObjectGlobal.QueryRow(sqlStatement, id)
	var bookToReturn Book
	err := foundRow.Scan(&bookToReturn.ID, &bookToReturn.Name, &bookToReturn.Price, &bookToReturn.Inventory, &bookToReturn.CreatedAt, &bookToReturn.UpdatedAt)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return Book{}, fmt.Errorf("searching by ID: %w", errBookNotFound)
		default:
			return Book{}, fmt.Errorf("searching by ID: %w", err)
		}
	}

	return bookToReturn, nil
}

/* Returns filtered content of database in a list of books*/
func listBooks(name string, minPrice32, maxPrice32 float32, sortBy, sortDirection string) ([]Book, error) {
	if name == "" {
		name = "%"
	}

	sqlStatement := fmt.Sprint(`SELECT * FROM bookstable 
	WHERE name LIKE $1
	AND price BETWEEN $2 AND $3	
	ORDER BY `, sortBy, ` `, sortDirection, ` ;`)

	rows, err := dbObjectGlobal.Query(sqlStatement, name, minPrice32, maxPrice32)
	if err != nil {
		return nil, fmt.Errorf("listing books from db: %w", err)
	}
	defer rows.Close()
	bookslist := []Book{}
	var bookToReturn Book
	for rows.Next() {
		err = rows.Scan(&bookToReturn.ID, &bookToReturn.Name, &bookToReturn.Price, &bookToReturn.Inventory, &bookToReturn.CreatedAt, &bookToReturn.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("listing books from db: %w", err)
		}

		bookslist = append(bookslist, bookToReturn)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("listing books from db: %w", err)
	}

	return bookslist, nil
}

/* Stores the book into the database, checks and returns it if succeed. */
func storeOnDB(newBook Book) (Book, error) {
	sqlStatement := `
	INSERT INTO bookstable (id, name, price, inventory, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING *`
	createdRow := dbObjectGlobal.QueryRow(sqlStatement, newBook.ID, newBook.Name, *newBook.Price, *newBook.Inventory, newBook.CreatedAt, newBook.UpdatedAt)
	var bookToReturn Book
	err := createdRow.Scan(&bookToReturn.ID, &bookToReturn.Name, &bookToReturn.Price, &bookToReturn.Inventory, &bookToReturn.CreatedAt, &bookToReturn.UpdatedAt)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return Book{}, fmt.Errorf("storing on db: %w", errBookNotFound)
		default:
			return Book{}, fmt.Errorf("storing on db: %w", err)
		}
	}

	return bookToReturn, nil
}
