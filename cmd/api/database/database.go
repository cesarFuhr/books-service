package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/books-service/cmd/api/book"
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
		return nil, fmt.Errorf("connecting to db, openning: %w", err)
	}

	err = sqlDB.Ping()
	if err != nil {
		return nil, fmt.Errorf("connecting to db, pingging: %w", err)
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

/* Change the status of 'archived' column on database. */
func (store *Store) SetBookArchiveStatus(ctx context.Context, id uuid.UUID, archived bool) (book.Book, error) {
	sqlStatement := `
	UPDATE bookstable 
	SET archived = $2
	WHERE id = $1
	RETURNING *`
	updatedRow := store.db.QueryRowContext(ctx, sqlStatement, id, archived)
	var bookToReturn book.Book
	err := updatedRow.Scan(&bookToReturn.ID, &bookToReturn.Name, &bookToReturn.Price, &bookToReturn.Inventory, &bookToReturn.CreatedAt, &bookToReturn.UpdatedAt, &bookToReturn.Archived)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return book.Book{}, fmt.Errorf("archiving on db: %w", book.ErrResponseBookNotFound)
		default:
			return book.Book{}, fmt.Errorf("archiving on db: %w", err)
		}
	}

	return bookToReturn, nil
}

/* Stores the book into the database, checks and returns it if succeed. */
func (store *Store) CreateBook(ctx context.Context, bookEntry book.Book) (book.Book, error) {
	sqlStatement := `
	INSERT INTO bookstable (id, name, price, inventory, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING *`
	createdRow := store.db.QueryRowContext(ctx, sqlStatement, bookEntry.ID, bookEntry.Name, *bookEntry.Price, *bookEntry.Inventory, bookEntry.CreatedAt, bookEntry.UpdatedAt)
	var bookToReturn book.Book
	err := createdRow.Scan(&bookToReturn.ID, &bookToReturn.Name, &bookToReturn.Price, &bookToReturn.Inventory, &bookToReturn.CreatedAt, &bookToReturn.UpdatedAt, &bookToReturn.Archived)
	if err != nil {
		return book.Book{}, fmt.Errorf("storing book on db: %w", err)
	}

	return bookToReturn, nil
}

/* Searches a book in database based on ID and returns it if succeed. */
func (store *Store) GetBookByID(ctx context.Context, id uuid.UUID) (book.Book, error) {
	sqlStatement := `SELECT id, name, price, inventory, created_at, updated_at, archived
	FROM bookstable 
	WHERE id=$1;`
	foundRow := store.db.QueryRowContext(ctx, sqlStatement, id)
	var bookToReturn book.Book
	err := foundRow.Scan(&bookToReturn.ID, &bookToReturn.Name, &bookToReturn.Price, &bookToReturn.Inventory, &bookToReturn.CreatedAt, &bookToReturn.UpdatedAt, &bookToReturn.Archived)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return book.Book{}, fmt.Errorf("searching by ID: %w", book.ErrResponseBookNotFound)
		default:
			return book.Book{}, fmt.Errorf("searching by ID: %w", err)
		}
	}

	return bookToReturn, nil
}

/* Returns filtered content of database in a list of books*/
func (store *Store) ListBooks(ctx context.Context, name string, minPrice32, maxPrice32 float32, sortBy, sortDirection string, archived bool, page, pageSize int) ([]book.Book, error) {
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

	rows, err := store.db.QueryContext(ctx, sqlStatement, name, minPrice32, maxPrice32, archived)
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

/* Stores the book into the database, checks and returns it if succeed. */
func (store *Store) UpdateBook(ctx context.Context, bookEntry book.Book) (book.Book, error) {
	sqlStatement := `
	UPDATE bookstable 
	SET name = $2, price = $3, inventory = $4, updated_at = $5
	WHERE id = $1
	RETURNING *`
	updatedRow := store.db.QueryRowContext(ctx, sqlStatement, bookEntry.ID, bookEntry.Name, *bookEntry.Price, *bookEntry.Inventory, bookEntry.UpdatedAt)
	var bookToReturn book.Book
	err := updatedRow.Scan(&bookToReturn.ID, &bookToReturn.Name, &bookToReturn.Price, &bookToReturn.Inventory, &bookToReturn.CreatedAt, &bookToReturn.UpdatedAt, &bookToReturn.Archived)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return book.Book{}, fmt.Errorf("updating on db: %w", book.ErrResponseBookNotFound)
		default:
			return book.Book{}, fmt.Errorf("updating on db: %w", err)
		}
	}

	return bookToReturn, nil
}

/* Counts how many rows in db fit the specified filter parameters. */
func (store *Store) ListBooksTotals(ctx context.Context, name string, minPrice32, maxPrice32 float32, archived bool) (int, error) {
	if name != "" {
		name = fmt.Sprint("%", name, "%")
	} else {
		name = "%"
	}

	sqlStatement := `SELECT COUNT(*) FROM bookstable 
	WHERE name ILIKE $1
	AND (archived = $4 OR archived = FALSE)
	AND price BETWEEN $2 AND $3;`

	row := store.db.QueryRowContext(ctx, sqlStatement, name, minPrice32, maxPrice32, archived)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return count, fmt.Errorf("counting books from db: %w", err)
	}

	return count, nil
}

/* Stores a new order into the database, checks and returns it if succeed. */
func (store *Store) CreateOrder(ctx context.Context, newOrder book.Order) (book.Order, error) {
	sqlStatement := `
	INSERT INTO orders (order_id, purchaser_id, order_status, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING *`
	createdRow := store.db.QueryRowContext(ctx, sqlStatement, newOrder.OrderID, newOrder.PurchaserID, newOrder.OrderStatus, newOrder.CreatedAt, newOrder.UpdatedAt)
	var orderToReturn book.Order
	err := createdRow.Scan(&orderToReturn.OrderID, &orderToReturn.PurchaserID, &orderToReturn.OrderStatus, &orderToReturn.CreatedAt, &orderToReturn.UpdatedAt)
	if err != nil {
		return book.Order{}, fmt.Errorf("storing order on db: %w", err)
	}

	return orderToReturn, nil
}

func (store *Store) ListOrderItems(ctx context.Context, order_id uuid.UUID) (book.Order, []book.OrderItem, error) {
	sqlStatement := `SELECT order_id, purchaser_id, order_status, created_at, updated_at
	FROM orders 
	WHERE order_id=$1;`
	foundRow := store.db.QueryRowContext(ctx, sqlStatement, order_id)
	var orderToReturn book.Order
	err := foundRow.Scan(&orderToReturn.OrderID, &orderToReturn.PurchaserID, &orderToReturn.OrderStatus, &orderToReturn.CreatedAt, &orderToReturn.UpdatedAt)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return book.Order{}, []book.OrderItem{}, fmt.Errorf("listing order items from db: %w", book.ErrResponseOrderNotFound)
		default:
			return book.Order{}, []book.OrderItem{}, fmt.Errorf("listing order items from db: %w", err)
		}
	}

	sqlStatement = `SELECT * FROM books_orders 
	WHERE order_id=$1;`

	rows, err := store.db.QueryContext(ctx, sqlStatement, order_id)
	if err != nil {
		return book.Order{}, []book.OrderItem{}, fmt.Errorf("listing order items from db: %w", err)
	}
	defer rows.Close()
	itemsAtOrderList := []book.OrderItem{}
	var itemAtOrder book.OrderItem
	for rows.Next() {
		err = rows.Scan(&itemAtOrder.OrderID, &itemAtOrder.BookID, &itemAtOrder.BookUnits, &itemAtOrder.BookPriceAtOrder, &itemAtOrder.CreatedAt, &itemAtOrder.UpdatedAt)
		if err != nil {
			return book.Order{}, []book.OrderItem{}, fmt.Errorf("listing order items from db: %w", err)
		}

		itemsAtOrderList = append(itemsAtOrderList, itemAtOrder)
	}

	err = rows.Err()
	if err != nil {
		return book.Order{}, []book.OrderItem{}, fmt.Errorf("listing order items from db: %w", err)
	}

	return orderToReturn, itemsAtOrderList, nil
}

func (store *Store) AddItemToOrder(ctx context.Context, newItemAtOrder book.OrderItem) (book.OrderItem, error) {
	sqlStatement := `
	INSERT INTO books_orders (order_id, book_id, book_units, book_price_at_order, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING *`
	createdRow := store.db.QueryRowContext(ctx, sqlStatement, newItemAtOrder.OrderID, newItemAtOrder.BookID, newItemAtOrder.BookUnits, *newItemAtOrder.BookPriceAtOrder, newItemAtOrder.CreatedAt, newItemAtOrder.UpdatedAt)
	var itemToReturn book.OrderItem
	err := createdRow.Scan(&itemToReturn.OrderID, &itemToReturn.BookID, &itemToReturn.BookUnits, &itemToReturn.BookPriceAtOrder, &itemToReturn.CreatedAt, &itemToReturn.UpdatedAt)
	if err != nil {
		return book.OrderItem{}, fmt.Errorf("storing new item at order on db: %w", err)
	}

	return itemToReturn, nil
}

/* Updates an order stored in database, adding or removing items(books) from it. */
func (store *Store) UpdateOrder(ctx context.Context, updtReq book.UpdateOrderRequest) (book.OrderItem, error) {
	var itemToReturn book.OrderItem

	_, err := store.db.ExecContext(ctx, `BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE;`)
	if err != nil {
		return book.OrderItem{}, fmt.Errorf("beginning transaction to update order on db: %w", err)
	}

	defer func() {
		_, err = store.db.ExecContext(ctx, `ROLLBACK;`)
		if err != nil {
			itemToReturn = book.OrderItem{}
			err = fmt.Errorf("rolling back transaction to update order on db: %w", err)
		}
	}()

	//WRITE TRANSACTION FUNCTIONS HERE:

	return itemToReturn, nil
}
