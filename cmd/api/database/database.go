package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log"
	"time"

	"github.com/books-service/cmd/api/book"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/google/uuid"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "github.com/lib/pq"
)

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type Store struct {
	db  *sql.DB
	exc *Exectuor
}

type Exectuor struct {
	DBTX
}

func NewStore(db *sql.DB) *Store {
	CurrentStore := &Store{
		db:  db,
		exc: NewExc(db),
	}
	return CurrentStore
}

func NewExc(dbtx DBTX) *Exectuor {
	return &Exectuor{DBTX: dbtx}
}

func (store *Store) BeginTx(ctx context.Context, opts *sql.TxOptions) (book.Repository, driver.Tx, error) {
	tx, err := store.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("beginning transaction: %w", err)
	}

	txRepo := NewStore(store.db)
	txRepo.exc = NewExc(tx)
	return txRepo, tx, nil
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
	updatedRow := store.exc.QueryRowContext(ctx, sqlStatement, id, archived)
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
	createdRow := store.exc.QueryRowContext(ctx, sqlStatement, bookEntry.ID, bookEntry.Name, *bookEntry.Price, *bookEntry.Inventory, bookEntry.CreatedAt, bookEntry.UpdatedAt)
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
	foundRow := store.exc.QueryRowContext(ctx, sqlStatement, id)
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

	rows, err := store.exc.QueryContext(ctx, sqlStatement, name, minPrice32, maxPrice32, archived)
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
	updatedRow := store.exc.QueryRowContext(ctx, sqlStatement, bookEntry.ID, bookEntry.Name, *bookEntry.Price, *bookEntry.Inventory, bookEntry.UpdatedAt)
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

	row := store.exc.QueryRowContext(ctx, sqlStatement, name, minPrice32, maxPrice32, archived)
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
	createdRow := store.exc.QueryRowContext(ctx, sqlStatement, newOrder.OrderID, newOrder.PurchaserID, newOrder.OrderStatus, newOrder.CreatedAt, newOrder.UpdatedAt)
	var orderToReturn book.Order
	err := createdRow.Scan(&orderToReturn.OrderID, &orderToReturn.PurchaserID, &orderToReturn.OrderStatus, &orderToReturn.CreatedAt, &orderToReturn.UpdatedAt)
	if err != nil {
		return book.Order{}, fmt.Errorf("storing order on db: %w", err)
	}

	return orderToReturn, nil
}

func (store *Store) ListOrderItems(ctx context.Context, order_id uuid.UUID) (book.Order, error) {
	sqlStatement := `SELECT order_id, purchaser_id, order_status, created_at, updated_at
	FROM orders 
	WHERE order_id=$1;`
	foundRow := store.exc.QueryRowContext(ctx, sqlStatement, order_id)
	var orderToReturn book.Order
	err := foundRow.Scan(&orderToReturn.OrderID, &orderToReturn.PurchaserID, &orderToReturn.OrderStatus, &orderToReturn.CreatedAt, &orderToReturn.UpdatedAt)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return book.Order{}, fmt.Errorf("listing order items from db: %w", book.ErrResponseOrderNotFound)
		default:
			return book.Order{}, fmt.Errorf("listing order items from db: %w", err)
		}
	}

	sqlStatement = `SELECT book_id, book_units, book_price_at_order, created_at, updated_at, book_name
	FROM books_orders 
	WHERE order_id=$1
	ORDER BY updated_at ASC;`

	rows, err := store.exc.QueryContext(ctx, sqlStatement, order_id)
	if err != nil {
		return book.Order{}, fmt.Errorf("listing order items from db: %w", err)
	}
	defer rows.Close()
	var itemAtOrder book.OrderItem
	for rows.Next() {
		err = rows.Scan(&itemAtOrder.BookID, &itemAtOrder.BookUnits, &itemAtOrder.BookPriceAtOrder, &itemAtOrder.CreatedAt, &itemAtOrder.UpdatedAt, &itemAtOrder.BookName)
		if err != nil {
			return book.Order{}, fmt.Errorf("listing order items from db: %w", err)
		}

		orderToReturn.Items = append(orderToReturn.Items, itemAtOrder)

		orderToReturn.TotalPrice = orderToReturn.TotalPrice + (*itemAtOrder.BookPriceAtOrder * float32(itemAtOrder.BookUnits))
	}

	err = rows.Err()
	if err != nil {
		return book.Order{}, fmt.Errorf("listing order items from db: %w", err)
	}

	return orderToReturn, nil
}

/* Updates a row in orders table and checks if the order is accepting items. */
func (store *Store) UpdateOrderRow(ctx context.Context, orderID uuid.UUID) error {
	sqlStatement := `
	UPDATE orders
	SET updated_at = $2
	WHERE order_id = $1
	RETURNING order_status`
	updatedRow := store.exc.QueryRowContext(ctx, sqlStatement, orderID, time.Now().UTC().Round(time.Millisecond))
	var o book.Order
	err := updatedRow.Scan(&o.OrderStatus)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return fmt.Errorf("updating order on db: %w", book.ErrResponseOrderNotFound)
		default:
			return fmt.Errorf("updating order on db: %w", err)
		}
	}
	if o.OrderStatus != "accepting_items" {
		return fmt.Errorf("updating order on db: %w", book.ErrResponseOrderNotAcceptingItems)
	}

	return nil
}

/*Gets a book from the order searching by ID */
func (store *Store) GetOrderItem(ctx context.Context, orderID uuid.UUID, bookID uuid.UUID) (book.OrderItem, error) {
	sqlStatement := `SELECT book_id, book_units, book_price_at_order, created_at, updated_at, book_name
	FROM books_orders 
	WHERE order_id=$1 AND book_id=$2;`
	foundRow := store.exc.QueryRowContext(ctx, sqlStatement, orderID, bookID)
	var itemToReturn book.OrderItem
	err := foundRow.Scan(&itemToReturn.BookID, &itemToReturn.BookUnits, &itemToReturn.BookPriceAtOrder, &itemToReturn.CreatedAt, &itemToReturn.UpdatedAt, &itemToReturn.BookName)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return book.OrderItem{}, fmt.Errorf("getting order item from db: %w", book.ErrResponseBookNotAtOrder)
		default:
			return book.OrderItem{}, fmt.Errorf("getting order item from db: %w", err)
		}
	}

	return itemToReturn, nil
}

/*Inserts a new book into the order and, if it is already there, updates it */
func (store *Store) UpsertOrderItem(ctx context.Context, orderID uuid.UUID, itemToUpdt book.OrderItem) (book.OrderItem, error) {
	sqlStatement := `
	INSERT INTO books_orders (order_id, book_id, book_units, book_price_at_order, created_at, updated_at, book_name)
	VALUES ($1, $2, $3, $4, $5, $5, $6)
	ON CONFLICT ON CONSTRAINT books_orders_pkey DO UPDATE
	SET book_units = $3, updated_at = $5
	WHERE books_orders.order_id=$1 AND books_orders.book_id=$2
	RETURNING book_id, book_units, book_price_at_order, created_at, updated_at, book_name`

	foundRow := store.exc.QueryRowContext(ctx, sqlStatement, orderID, itemToUpdt.BookID, itemToUpdt.BookUnits, *itemToUpdt.BookPriceAtOrder, time.Now().UTC().Round(time.Millisecond), itemToUpdt.BookName)
	var itemToReturn book.OrderItem
	err := foundRow.Scan(&itemToReturn.BookID, &itemToReturn.BookUnits, &itemToReturn.BookPriceAtOrder, &itemToReturn.CreatedAt, &itemToReturn.UpdatedAt, &itemToReturn.BookName)
	if err != nil {
		return book.OrderItem{}, fmt.Errorf("upserting item at order on db: %w", err)
	}
	return itemToReturn, nil
}

func (store *Store) DeleteOrderItem(ctx context.Context, orderID uuid.UUID, bookID uuid.UUID) error {
	sqlStatement := `
DELETE FROM books_orders
WHERE order_id = $1 AND book_id = $2;`
	_, err := store.exc.ExecContext(ctx, sqlStatement, orderID, bookID)
	if err != nil {
		return fmt.Errorf("deleting item from order on db: %w", err)
	}
	return nil
}
