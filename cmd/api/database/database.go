package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/books-service/cmd/api/book"
	bookerrors "github.com/books-service/cmd/api/errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/google/uuid"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "github.com/lib/pq"
)

type Store struct {
	db *sql.DB
}

/* Connects to the database trought a connection string and returns a pointer to a valid DB object (*sql.DB). */
func ConnectDb(connStr string) (*sql.DB, error) {

	sqlDB, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("connecting to db: %w", err)
	}

	err = sqlDB.Ping()
	if err != nil {
		return nil, fmt.Errorf("connecting to db: %w", err)
	}

	log.Println("Successfully connected!")
	return sqlDB, nil
}

func NewStore(db *sql.DB) *Store {
	CurrentStore := &Store{db: db}
	return CurrentStore
}

func MigrationUp(store *Store, path string) error {
	driver, err := postgres.WithInstance(store.db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migrating up: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", path),
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("migrating up: %w", err)
	}

	err = m.Up()
	if err != nil {
		return fmt.Errorf("migrating up: %w", err)
	}
	return nil
}

func (store *Store) CountRows(name string, minPrice32, maxPrice32 float32, archived bool) (int, error) {
	if name != "" {
		name = fmt.Sprint("%", name, "%")
	} else {
		name = "%"
	}

	sqlStatement := `SELECT COUNT(*) FROM bookstable 
	WHERE name ILIKE $1
	AND (archived = $4 OR archived = FALSE)
	AND price BETWEEN $2 AND $3;`

	row := store.db.QueryRow(sqlStatement, name, minPrice32, maxPrice32, archived)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return count, fmt.Errorf("counting books from db: %w", err)
	}

	return count, nil
}

/* Returns filtered content of database in a list of books*/
func (store *Store) ListBooks(name string, minPrice32, maxPrice32 float32, sortBy, sortDirection string, archived bool, page, pageSize int) ([]book.Book, error) {
	if name != "" {
		name = fmt.Sprint("%", name, "%")
	} else {
		name = "%"
	}

	limit := pageSize
	offset := (page - 1) * pageSize

	sqlStatement := fmt.Sprint(`SELECT * FROM bookstable 
	WHERE name ILIKE $1
	AND (archived = $4 OR archived = FALSE)
	AND price BETWEEN $2 AND $3	
	ORDER BY `, sortBy, ` `, sortDirection, ` 
	LIMIT `, limit, ` OFFSET `, offset, ` ;`)

	rows, err := store.db.Query(sqlStatement, name, minPrice32, maxPrice32, archived)
	if err != nil {
		return nil, fmt.Errorf("listing books from db: %w", err)
	}
	defer rows.Close()
	bookslist := []book.Book{}
	var bookToReturn book.Book
	for rows.Next() {
		err = rows.Scan(&bookToReturn.ID, &bookToReturn.Name, &bookToReturn.Price, &bookToReturn.Inventory, &bookToReturn.CreatedAt, &bookToReturn.UpdatedAt, &bookToReturn.Archived)
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

//==========BOOK STORING FUNCTIONS:===========

/* Change the status of 'archived' column on database. */
func (store *Store) ArchiveStatusBook(id uuid.UUID, archived bool) (book.Book, error) {
	sqlStatement := `
	UPDATE bookstable 
	SET archived = $2
	WHERE id = $1
	RETURNING *`
	updatedRow := store.db.QueryRow(sqlStatement, id, archived)
	var bookToReturn book.Book
	err := updatedRow.Scan(&bookToReturn.ID, &bookToReturn.Name, &bookToReturn.Price, &bookToReturn.Inventory, &bookToReturn.CreatedAt, &bookToReturn.UpdatedAt, &bookToReturn.Archived)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return book.Book{}, fmt.Errorf("archiving on db: %w", bookerrors.ErrResponseBookNotFound)
		default:
			return book.Book{}, fmt.Errorf("archiving on db: %w", err)
		}
	}

	return bookToReturn, nil
}

/* Stores the book into the database, checks and returns it if succeed. */
func (store *Store) CreateBook(bookEntry book.Book) (book.Book, error) {
	sqlStatement := `
	INSERT INTO bookstable (id, name, price, inventory, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING *`
	createdRow := store.db.QueryRow(sqlStatement, bookEntry.ID, bookEntry.Name, *bookEntry.Price, *bookEntry.Inventory, bookEntry.CreatedAt, bookEntry.UpdatedAt)
	var bookToReturn book.Book
	err := createdRow.Scan(&bookToReturn.ID, &bookToReturn.Name, &bookToReturn.Price, &bookToReturn.Inventory, &bookToReturn.CreatedAt, &bookToReturn.UpdatedAt, &bookToReturn.Archived)
	if err != nil {
		return book.Book{}, fmt.Errorf("storing on db: %w", err)
	}

	return bookToReturn, nil
}

/* Searches a book in database based on ID and returns it if succeed. */
func (store *Store) GetBookByID(id uuid.UUID) (book.Book, error) {
	sqlStatement := `SELECT id, name, price, inventory, created_at, updated_at, archived FROM bookstable WHERE id=$1;`
	foundRow := store.db.QueryRow(sqlStatement, id)
	var bookToReturn book.Book
	err := foundRow.Scan(&bookToReturn.ID, &bookToReturn.Name, &bookToReturn.Price, &bookToReturn.Inventory, &bookToReturn.CreatedAt, &bookToReturn.UpdatedAt, &bookToReturn.Archived)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return book.Book{}, fmt.Errorf("searching by ID: %w", bookerrors.ErrResponseBookNotFound)
		default:
			return book.Book{}, fmt.Errorf("searching by ID: %w", err)
		}
	}

	return bookToReturn, nil
}

/* Stores the book into the database, checks and returns it if succeed. */
func (store *Store) UpdateBook(bookEntry book.Book) (book.Book, error) {
	sqlStatement := `
	UPDATE bookstable 
	SET name = $2, price = $3, inventory = $4, updated_at = $5
	WHERE id = $1
	RETURNING *`
	updatedRow := store.db.QueryRow(sqlStatement, bookEntry.ID, bookEntry.Name, *bookEntry.Price, *bookEntry.Inventory, bookEntry.UpdatedAt)
	var bookToReturn book.Book
	err := updatedRow.Scan(&bookToReturn.ID, &bookToReturn.Name, &bookToReturn.Price, &bookToReturn.Inventory, &bookToReturn.CreatedAt, &bookToReturn.UpdatedAt, &bookToReturn.Archived)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return book.Book{}, fmt.Errorf("updating on db: %w", bookerrors.ErrResponseBookNotFound)
		default:
			return book.Book{}, fmt.Errorf("updating on db: %w", err)
		}
	}

	return bookToReturn, nil
}
