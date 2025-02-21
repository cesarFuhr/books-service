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
	db  *memdb.MemDB
	exc *memdb.Txn
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
			"books_orders": {
				Name: "books_orders",
				Indexes: map[string]*memdb.IndexSchema{
					"id": { // Composite index for quick lookups
						Name:   "id",
						Unique: true,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "OrderID"},
								&memdb.StringFieldIndex{Field: "BookID"},
							},
						},
					},
					"order_id": {
						Name:    "order_id",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "OrderID"},
					},
					"book_id": {
						Name:    "book_id",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "BookID"},
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
	return &InMemoryStore{db: db, exc: nil}, nil
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

type AdaptedOrder struct {
	OrderID     string
	PurchaserID string
	OrderStatus string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	TotalPrice  float32
	Items       []book.OrderItem
}

func adaptOrderIdToString(o book.Order) AdaptedOrder {
	return AdaptedOrder{
		OrderID:     o.OrderID.String(),
		PurchaserID: o.PurchaserID.String(),
		OrderStatus: o.OrderStatus,
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
		TotalPrice:  o.TotalPrice,
		Items:       o.Items,
	}
}

func adaptOrderIdToUUID(adptOrder AdaptedOrder) book.Order {
	return book.Order{
		OrderID:     uuid.MustParse(adptOrder.OrderID),
		PurchaserID: uuid.MustParse(adptOrder.PurchaserID),
		OrderStatus: adptOrder.OrderStatus,
		CreatedAt:   adptOrder.CreatedAt,
		UpdatedAt:   adptOrder.UpdatedAt,
		TotalPrice:  adptOrder.TotalPrice,
		Items:       adptOrder.Items,
	}
}

type AdaptedOrderItem struct {
	OrderID          string
	BookID           string
	BookName         string
	BookUnits        int
	BookPriceAtOrder *float32
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func adaptOrderItemIdToString(orderID uuid.UUID, o book.OrderItem) AdaptedOrderItem {
	return AdaptedOrderItem{
		OrderID:          orderID.String(),
		BookID:           o.BookID.String(),
		BookName:         o.BookName,
		BookUnits:        o.BookUnits,
		BookPriceAtOrder: o.BookPriceAtOrder,
		CreatedAt:        o.CreatedAt,
		UpdatedAt:        o.UpdatedAt,
	}
}

func adaptOrderItemIdToUUID(o AdaptedOrderItem) book.OrderItem {
	return book.OrderItem{
		BookID:           uuid.MustParse(o.BookID),
		BookName:         o.BookName,
		BookUnits:        o.BookUnits,
		BookPriceAtOrder: o.BookPriceAtOrder,
		CreatedAt:        o.CreatedAt,
		UpdatedAt:        o.UpdatedAt,
	}
}

// -- Books --

func (store *InMemoryStore) SetBookArchiveStatus(ctx context.Context, id uuid.UUID, archived bool) (book.Book, error) {
	insideTx := true
	if store.exc == nil { //It means this method is not being called inside a larger transction.
		insideTx = false
		store.exc = store.db.Txn(true)
		defer store.endTX()
	}

	raw, err := store.exc.First("book", "id", id.String())
	if err != nil {
		return book.Book{}, fmt.Errorf("archiving on db: %w", err)
	}
	if raw == nil {
		return book.Book{}, fmt.Errorf("archiving on db: %w", book.ErrResponseBookNotFound)
	}

	updatedBook := raw.(AdaptedBook)
	updatedBook.Archived = archived

	if err := store.exc.Insert("book", updatedBook); err != nil {
		return book.Book{}, err
	}

	if !insideTx {
		store.exc.Commit()
	}
	return adaptBookIdToUUID(updatedBook), nil
}

func (store *InMemoryStore) CreateBook(ctx context.Context, bookEntry book.Book) (book.Book, error) {
	insideTx := true
	if store.exc == nil { //It means this method is not being called inside a larger transction.
		insideTx = false
		store.exc = store.db.Txn(true)
		defer store.endTX()
	}

	err := store.exc.Insert("book", adaptBookIdToString(bookEntry))
	if err != nil {
		return book.Book{}, fmt.Errorf("storing book on db: %w", err)
	}

	raw, err := store.exc.First("book", "id", bookEntry.ID.String())
	if err != nil {
		return book.Book{}, fmt.Errorf("storing book on db: %w", err)
	}
	if raw == nil {
		return book.Book{}, fmt.Errorf("storing book on db: %w", book.ErrResponseBookNotFound)
	}

	if !insideTx {
		store.exc.Commit()
	}

	return adaptBookIdToUUID(raw.(AdaptedBook)), nil
}

func (store *InMemoryStore) GetBookByID(ctx context.Context, id uuid.UUID) (book.Book, error) {
	//insideTx := true
	if store.exc == nil { //It means this method is not being called inside a larger transction.
		//	insideTx = false
		store.exc = store.db.Txn(false)
		defer store.endTX()
	}

	raw, err := store.exc.First("book", "id", id.String())
	if err != nil {
		return book.Book{}, fmt.Errorf("searching by ID: %w", err)
	}
	if raw == nil {
		return book.Book{}, fmt.Errorf("searching by ID: %w", book.ErrResponseBookNotFound)
	}

	return adaptBookIdToUUID(raw.(AdaptedBook)), nil
}

func (store *InMemoryStore) ListBooks(ctx context.Context, name string, minPrice32, maxPrice32 float32, sortBy, sortDirection string, archived bool, page, pageSize int) ([]book.Book, error) {
	//insideTx := true
	if store.exc == nil { //It means this method is not being called inside a larger transction.
		//	insideTx = false
		store.exc = store.db.Txn(false)
		defer store.endTX()
	}

	it, err := store.exc.Get("book", "id")
	if err != nil {
		return []book.Book{}, fmt.Errorf("listing books from db: %w", err)
	}

	books := []book.Book{}
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

	if len(books) > 1 {
		booksSorted := sortBooks(sortBy, sortDirection, books)

		// Apply pagination
		start := (page - 1) * pageSize

		end := start + pageSize
		if end > len(books) {
			end = len(books)
		}

		return booksSorted[start:end], nil
	}

	return books, nil
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
	//insideTx := true
	if store.exc == nil { //It means this method is not being called inside a larger transction.
		//	insideTx = false
		store.exc = store.db.Txn(false)
		defer store.endTX()
	}

	it, err := store.exc.Get("book", "id")
	if err != nil {
		return 0, fmt.Errorf("counting books from db: %w", err)
	}

	books := []book.Book{}
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
	insideTx := true
	if store.exc == nil { //It means this method is not being called inside a larger transction.
		insideTx = false
		store.exc = store.db.Txn(true)
		defer store.endTX()
	}

	raw, err := store.exc.First("book", "id", bookEntry.ID.String())
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

	if err := store.exc.Insert("book", updatedBook); err != nil {
		return book.Book{}, err
	}

	if !insideTx {
		store.exc.Commit()
	}
	return bookEntry, nil
}

// -- Orders --

func (store *InMemoryStore) CreateOrder(ctx context.Context, newOrder book.Order) (book.Order, error) {
	insideTx := true
	if store.exc == nil { //It means this method is not being called inside a larger transction.
		insideTx = false
		store.exc = store.db.Txn(true)
		defer store.endTX()
	}

	if err := store.exc.Insert("order", adaptOrderIdToString(newOrder)); err != nil {
		return book.Order{}, fmt.Errorf("storing order on db: %w", err)
	}

	if !insideTx {
		store.exc.Commit()
	}
	return newOrder, nil
}

func (store *InMemoryStore) ListOrderItems(ctx context.Context, orderID uuid.UUID) (book.Order, error) {
	//insideTx := true
	if store.exc == nil { //It means this method is not being called inside a larger transction.
		//	insideTx = false
		store.exc = store.db.Txn(false)
		defer store.endTX()
	}

	raw, err := store.exc.First("order", "id", orderID.String())
	if err != nil {
		return book.Order{}, fmt.Errorf("listing order items from db: %w", err)
	}
	if raw == nil {
		return book.Order{}, fmt.Errorf("listing order items from db: %w", book.ErrResponseOrderNotFound)
	}
	orderToReturn := adaptOrderIdToUUID(raw.(AdaptedOrder))

	it, err := store.exc.Get("books_orders", "order_id", orderID.String())
	if err != nil {
		return book.Order{}, fmt.Errorf("listing order items from db: %w", err)
	}

	items := []book.OrderItem{}
	for obj := it.Next(); obj != nil; obj = it.Next() {
		b := obj.(AdaptedOrderItem)
		items = append(items, adaptOrderItemIdToUUID(b))
		orderToReturn.TotalPrice = orderToReturn.TotalPrice + (*b.BookPriceAtOrder * float32(b.BookUnits))
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].UpdatedAt.Before(items[j].UpdatedAt)
	})

	orderToReturn.Items = items

	return orderToReturn, nil
}

func (store *InMemoryStore) GetOrderItem(ctx context.Context, orderID uuid.UUID, bookID uuid.UUID) (book.OrderItem, error) {
	//insideTx := true
	if store.exc == nil { //It means this method is not being called inside a larger transction.
		//	insideTx = false
		store.exc = store.db.Txn(false)
		defer store.endTX()
	}

	raw, err := store.exc.First("books_orders", "id", orderID.String(), bookID.String())
	if err != nil {
		return book.OrderItem{}, fmt.Errorf("getting order item from db: %w", err)
	}
	if raw == nil {
		return book.OrderItem{}, fmt.Errorf("getting order item from db: %w", book.ErrResponseBookNotAtOrder)
	}

	return adaptOrderItemIdToUUID(raw.(AdaptedOrderItem)), nil
}

func (store *InMemoryStore) UpdateOrderRow(ctx context.Context, orderID uuid.UUID) error {
	insideTx := true
	if store.exc == nil { //It means this method is not being called inside a larger transction.
		insideTx = false
		store.exc = store.db.Txn(true)
		defer store.endTX()
	}

	raw, err := store.exc.First("order", "id", orderID.String())
	if err != nil {
		return fmt.Errorf("updating order on db: %w", err)
	}
	if raw == nil {
		return fmt.Errorf("updating order on db: %w", book.ErrResponseOrderNotFound)
	}

	order := raw.(AdaptedOrder)
	if order.OrderStatus != "accepting_items" {
		return fmt.Errorf("updating order on db: %w", book.ErrResponseOrderNotAcceptingItems)
	}

	order.UpdatedAt = time.Now().UTC().Round(time.Millisecond)
	if err := store.exc.Insert("order", order); err != nil {
		return fmt.Errorf("updating order on db: %w", err)
	}

	if !insideTx {
		store.exc.Commit()
	}
	return nil
}

func (store *InMemoryStore) UpsertOrderItem(ctx context.Context, orderID uuid.UUID, itemToUpdt book.OrderItem) (book.OrderItem, error) {
	insideTx := true
	if store.exc == nil { //It means this method is not being called inside a larger transction.
		insideTx = false
		store.exc = store.db.Txn(true)
		defer store.endTX()
	}

	raw, err := store.exc.First("books_orders", "id", orderID.String(), itemToUpdt.BookID.String())
	if err != nil {
		return book.OrderItem{}, fmt.Errorf("upserting item at order on db: %w", err)
	}
	orderItem := AdaptedOrderItem{}
	if raw == nil {
		orderItem = adaptOrderItemIdToString(orderID, itemToUpdt)
		orderItem.CreatedAt = time.Now().UTC().Round(time.Millisecond)
		orderItem.UpdatedAt = time.Now().UTC().Round(time.Millisecond)
	} else {
		orderItem = raw.(AdaptedOrderItem)
		orderItem.BookUnits = itemToUpdt.BookUnits
		orderItem.UpdatedAt = time.Now().UTC().Round(time.Millisecond)
	}

	if err := store.exc.Insert("books_orders", orderItem); err != nil {
		return book.OrderItem{}, fmt.Errorf("upserting item at order on db: %w", err)
	}

	if !insideTx {
		store.exc.Commit()
	}
	return adaptOrderItemIdToUUID(orderItem), nil
}

func (store *InMemoryStore) DeleteOrderItem(ctx context.Context, orderID uuid.UUID, bookID uuid.UUID) error {
	insideTx := true
	if store.exc == nil { //It means this method is not being called inside a larger transction.
		insideTx = false
		store.exc = store.db.Txn(true)
		defer store.endTX()
	}

	count, err := store.exc.DeleteAll("books_orders", "id", orderID.String(), bookID.String())
	if err != nil && count != 1 {
		return fmt.Errorf("deleting item from order on db: %w", err)
	}

	if !insideTx {
		store.exc.Commit()
	}
	return nil
}

// -- Transactions --

func (store *InMemoryStore) BeginTx(ctx context.Context, opts *sql.TxOptions) (book.Repository, driver.Tx, error) {
	txn := store.db.Txn(true)
	if txn == nil {
		return nil, nil, fmt.Errorf("failed to create transaction")
	}

	txWrapper := &TxWrapper{txn: txn}
	txStore := &InMemoryStore{
		db:  store.db,
		exc: txWrapper.txn,
	}

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
	var nilTxn *memdb.Txn
	tx.txn = nilTxn
	return nil
}

func (store *InMemoryStore) endTX() {
	store.exc.Abort()
	var nilTxn *memdb.Txn
	store.exc = nilTxn
}
