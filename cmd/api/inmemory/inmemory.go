package inmemory

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/books-service/cmd/api/book"
	"github.com/google/uuid"
	"github.com/hashicorp/go-memdb"
)

type InMemoryStore struct {
	db *memdb.MemDB
}

func NewInMemoryStore() (*InMemoryStore, error) {
	// Define the schema
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"book": {
				Name: "book",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
				},
			},
			"order": {
				Name: "order",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "OrderID"},
					},
				},
			},
		},
	}

	errV := schema.Validate()
	if errV != nil {
		log.Println("schema validating error: ", errV)
	}

	db, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize in-memory database: %w", err)
	}

	return &InMemoryStore{db: db}, nil
}

type AdaptedBook struct {
	ID        string
	Name      string
	Price     *float32
	Inventory *int
	CreatedAt time.Time
	UpdatedAt time.Time
	Archived  bool
}

func adaptBookIdToString(bookEntry book.Book) AdaptedBook {
	return AdaptedBook{
		ID:        bookEntry.ID.String(),
		Name:      bookEntry.Name,
		Price:     bookEntry.Price,
		Inventory: bookEntry.Inventory,
		CreatedAt: bookEntry.CreatedAt,
		UpdatedAt: bookEntry.UpdatedAt,
		Archived:  bookEntry.Archived,
	}
}

func adaptBookIdToUUID(adptBook AdaptedBook) book.Book {
	return book.Book{
		ID:        uuid.MustParse(adptBook.ID),
		Name:      adptBook.Name,
		Price:     adptBook.Price,
		Inventory: adptBook.Inventory,
		CreatedAt: adptBook.CreatedAt,
		UpdatedAt: adptBook.UpdatedAt,
		Archived:  adptBook.Archived,
	}
}

// -- Books --

func (store *InMemoryStore) SetBookArchiveStatus(ctx context.Context, id uuid.UUID, archived bool) (book.Book, error) {
	txn := store.db.Txn(true)
	defer txn.Abort()

	raw, err := txn.First("book", "id", id.String())
	if err != nil {
		return book.Book{}, fmt.Errorf("archiving on db: %w", err)
	}
	if raw == nil {
		return book.Book{}, fmt.Errorf("archiving on db: %w", book.ErrResponseBookNotFound)
	}

	updatedBook := raw.(AdaptedBook)
	updatedBook.Archived = archived
	updatedBook.UpdatedAt = time.Now()

	if err := txn.Insert("book", updatedBook); err != nil {
		return book.Book{}, err
	}

	txn.Commit()
	return adaptBookIdToUUID(updatedBook), nil
}

func (store *InMemoryStore) CreateBook(ctx context.Context, bookEntry book.Book) (book.Book, error) {
	txn := store.db.Txn(true)
	defer txn.Abort()

	err := txn.Insert("book", adaptBookIdToString(bookEntry))
	if err != nil {
		return book.Book{}, fmt.Errorf("storing book on db: %w", err)
	}

	txn.Commit()
	return bookEntry, nil
}

func (store *InMemoryStore) GetBookByID(ctx context.Context, id uuid.UUID) (book.Book, error) {
	txn := store.db.Txn(false)
	defer txn.Abort()

	raw, err := txn.First("book", "id", id.String())
	if err != nil {
		return book.Book{}, fmt.Errorf("searching by ID: %w", err)
	}
	if raw == nil {
		return book.Book{}, fmt.Errorf("searching by ID: %w", book.ErrResponseBookNotFound)
	}

	return adaptBookIdToUUID(raw.(AdaptedBook)), nil
}

func (store *InMemoryStore) ListBooks(ctx context.Context, name string, minPrice32, maxPrice32 float32, sortBy, sortDirection string, archived bool, page, pageSize int) ([]book.Book, error) {
	txn := store.db.Txn(false)
	defer txn.Abort()

	it, err := txn.Get("book", "id")
	if err != nil {
		return nil, fmt.Errorf("listing books from db: %w", err)
	}

	var books []book.Book
	for obj := it.Next(); obj != nil; obj = it.Next() {
		b := obj.(AdaptedBook)
		if b.Archived != archived && b.Archived {
			continue //It only falls here if the 'b.Archived'==true and 'archived'==false, which means that the  query is not looking for archived books.
		}
		if name != "" && !strings.Contains(b.Name, name) {
			continue
		}
		if *b.Price < minPrice32 || *b.Price > maxPrice32 {
			continue
		}
		books = append(books, adaptBookIdToUUID(b))
	}

	booksSorted := sortBooks(sortBy, sortDirection, books)

	// Apply pagination
	start := (page - 1) * pageSize

	end := start + pageSize
	if end > len(books) {
		end = len(books)
	}

	return booksSorted[start:end], nil
}

func sortBooks(sortBy, sortDirection string, books []book.Book) []book.Book {
	if sortBy != "" {
		sort.Slice(books, func(i, j int) bool {
			switch sortBy {
			case "name":
				if sortDirection == "desc" {
					return books[i].Name > books[j].Name
				}
				return books[i].Name < books[j].Name
			case "price":
				if sortDirection == "desc" {
					return *books[i].Price > *books[j].Price
				}
				return *books[i].Price < *books[j].Price
			case "created_at":
				if sortDirection == "desc" {
					return books[i].CreatedAt.After(books[j].CreatedAt)
				}
				return books[i].CreatedAt.Before(books[j].CreatedAt)
			case "updated_at":
				if sortDirection == "desc" {
					return books[i].UpdatedAt.After(books[j].UpdatedAt)
				}
				return books[i].UpdatedAt.Before(books[j].UpdatedAt)
			case "inventory":
				if sortDirection == "desc" {
					return *books[i].Inventory > *books[j].Inventory
				}
				return *books[i].Inventory < *books[j].Inventory
			default:
				return true // No sorting applied
			}
		})
	}
	return books
}

func (store *InMemoryStore) ListBooksTotals(ctx context.Context, name string, minPrice32, maxPrice32 float32, archived bool) (int, error) {
	txn := store.db.Txn(false)
	defer txn.Abort()

	it, err := txn.Get("book", "id")
	if err != nil {
		return 0, fmt.Errorf("counting books from db: %w", err)
	}

	var books []book.Book
	for obj := it.Next(); obj != nil; obj = it.Next() {
		b := obj.(AdaptedBook)
		if b.Archived != archived && b.Archived {
			continue //It only falls here if the 'b.Archived'==true and 'archived'==false, which means that the  query is not looking for archived books.
		}
		if name != "" && !strings.Contains(b.Name, name) {
			continue
		}
		if *b.Price < minPrice32 || *b.Price > maxPrice32 {
			continue
		}
		books = append(books, adaptBookIdToUUID(b))
	}

	return len(books), nil
}

func (store *InMemoryStore) UpdateBook(ctx context.Context, bookEntry book.Book) (book.Book, error) {
	txn := store.db.Txn(true)
	defer txn.Abort()

	raw, err := txn.First("book", "id", bookEntry.ID.String())
	if err != nil {
		return book.Book{}, fmt.Errorf("updating book on db: %w", err)
	}
	if raw == nil {
		return book.Book{}, fmt.Errorf("updating book on db: %w", book.ErrResponseBookNotFound)
	}

	updatedBook := raw.(AdaptedBook)
	updatedBook.Name = bookEntry.Name
	updatedBook.Price = bookEntry.Price
	updatedBook.Inventory = bookEntry.Inventory
	//CreatedAt will not change
	updatedBook.UpdatedAt = bookEntry.UpdatedAt
	//Archived will not change

	if err := txn.Insert("book", updatedBook); err != nil {
		return book.Book{}, err
	}

	txn.Commit()
	return bookEntry, nil
}

// -- Orders --

func (store *InMemoryStore) CreateOrder(ctx context.Context, newOrder book.Order) (book.Order, error) {
	txn := store.db.Txn(true)
	defer txn.Abort()

	if err := txn.Insert("order", newOrder); err != nil {
		return book.Order{}, fmt.Errorf("storing order on db: %w", err)
	}

	txn.Commit()
	return newOrder, nil
}

func (store *InMemoryStore) ListOrderItems(ctx context.Context, orderID uuid.UUID) (book.Order, error) {
	txn := store.db.Txn(false)
	defer txn.Abort()

	raw, err := txn.First("order", "id", orderID)
	if err != nil {
		return book.Order{}, fmt.Errorf("listing order items from db: %w", err)
	}
	if raw == nil {
		return book.Order{}, fmt.Errorf("listing order items from db: %w", book.ErrResponseOrderNotFound)
	}

	return raw.(book.Order), nil
}

func (store *InMemoryStore) GetOrderItem(ctx context.Context, orderID uuid.UUID, bookID uuid.UUID) (book.OrderItem, error) {
	order, err := store.ListOrderItems(ctx, orderID)
	if err != nil {
		return book.OrderItem{}, fmt.Errorf("getting order item from db: %w", err)
	}

	for _, item := range order.Items {
		if item.BookID == bookID {
			return item, nil
		}
	}

	return book.OrderItem{}, fmt.Errorf("getting order item from db: %w", book.ErrResponseBookNotAtOrder)
}

func (store *InMemoryStore) UpdateOrderRow(ctx context.Context, orderID uuid.UUID) error {
	txn := store.db.Txn(true)
	defer txn.Abort()

	raw, err := txn.First("order", "id", orderID)
	if err != nil {
		return fmt.Errorf("updating order on db: %w", err)
	}
	if raw == nil {
		return fmt.Errorf("updating order on db: %w", book.ErrResponseOrderNotFound)
	}

	order := raw.(book.Order)
	if order.OrderStatus != "accepting_items" {
		return fmt.Errorf("updating order on db: %w", book.ErrResponseOrderNotAcceptingItems)
	}

	order.UpdatedAt = time.Now()
	if err := txn.Insert("order", order); err != nil {
		return fmt.Errorf("updating order on db: %w", err)
	}

	txn.Commit()
	return nil
}

func (store *InMemoryStore) UpsertOrderItem(ctx context.Context, orderID uuid.UUID, itemToUpdt book.OrderItem) (book.OrderItem, error) {
	txn := store.db.Txn(true)
	defer txn.Abort()

	raw, err := txn.First("order", "id", orderID)
	if err != nil {
		return book.OrderItem{}, fmt.Errorf("upserting item at order on db: %w", err)
	}
	if raw == nil {
		return book.OrderItem{}, fmt.Errorf("upserting item at order on db: %w", book.ErrResponseOrderNotFound)
	}

	order := raw.(book.Order)
	found := false
	for i, item := range order.Items {
		if item.BookID == itemToUpdt.BookID {
			order.Items[i] = itemToUpdt
			found = true
			break
		}
	}

	if !found {
		order.Items = append(order.Items, itemToUpdt)
	}

	if err := txn.Insert("order", order); err != nil {
		return book.OrderItem{}, fmt.Errorf("upserting item at order on db: %w", err)
	}

	txn.Commit()
	return itemToUpdt, nil
}

func (store *InMemoryStore) DeleteOrderItem(ctx context.Context, orderID uuid.UUID, bookID uuid.UUID) error {
	txn := store.db.Txn(true)
	defer txn.Abort()

	raw, err := txn.First("order", "id", orderID)
	if err != nil {
		return fmt.Errorf("deleting item from order on db: %w", err)
	}
	if raw == nil {
		return fmt.Errorf("deleting item from order on db: %w", book.ErrResponseOrderNotFound)
	}

	order := raw.(book.Order)
	newItems := []book.OrderItem{}
	for _, item := range order.Items {
		if item.BookID != bookID {
			newItems = append(newItems, item)
		}
	}
	order.Items = newItems

	if err := txn.Insert("order", order); err != nil {
		return fmt.Errorf("deleting item from order on db: %w", err)
	}

	txn.Commit()
	return nil
}

// -- Transactions --

func (store *InMemoryStore) BeginTx(ctx context.Context, opts *sql.TxOptions) (book.Repository, driver.Tx, error) {
	txn := store.db.Txn(true)
	if txn == nil {
		return nil, nil, fmt.Errorf("failed to create transaction")
	}

	txStore := &InMemoryStore{db: store.db}
	txWrapper := &TxWrapper{txn: txn}
	return txStore, txWrapper, nil
}

type TxWrapper struct {
	txn *memdb.Txn
}

func (tx *TxWrapper) Commit() error {
	tx.txn.Commit()
	return nil
}

func (tx *TxWrapper) Rollback() error {
	tx.txn.Abort()
	return nil
}
