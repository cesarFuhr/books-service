package database_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/books-service/cmd/api/book"
	"github.com/books-service/cmd/api/database"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/google/uuid"
	"github.com/matryer/is"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "github.com/lib/pq"
)

var store *database.Store
var sqlDB *sql.DB
var ctx context.Context = context.Background()

// TestMain is called before all the tests run.
// Usually is where we add logic to initialise resources.
func TestMain(m *testing.M) {
	// Setting up the database for tests.
	var err error
	connStr := os.Getenv("DATABASE_URL")
	sqlDB, err = database.ConnectDb(connStr)
	if err != nil {
		log.Fatalln(err)
	}

	store = database.NewStore(sqlDB)
	path := os.Getenv("DATABASE_MIGRATIONS_PATH")
	err = database.MigrationUp(store, path)
	if err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalln(err)
		}
		log.Println(err)
	}

	os.Exit(m.Run())
}

func TestCreateBook(t *testing.T) {
	// Removing all data from the test database.
	// We don't want to the database to be tainted with
	// this test data in another tests.
	t.Cleanup(func() {
		teardownDB(t)
	})

	t.Run("creates a book without errors", func(t *testing.T) {
		is := is.New(t)

		b := book.Book{
			ID:        uuid.New(),
			Name:      "A new book`",
			Price:     toPointer(float32(40.0)),
			Inventory: toPointer(10),
			CreatedAt: time.Now().UTC().Round(time.Millisecond),
			UpdatedAt: time.Now().UTC().Round(time.Millisecond),
		}

		newBook, err := store.CreateBook(ctx, b)
		is.NoErr(err)
		compareBooks(is, newBook, b)
	})
}
func TestArchiveStatusBook(t *testing.T) {
	t.Cleanup(func() {
		teardownDB(t)
	})

	t.Run("archives a book without errors", func(t *testing.T) {
		is := is.New(t)

		// Setting up, creating a book to be fetched.
		b := book.Book{
			ID:        uuid.New(),
			Name:      "A new book to be archived",
			Price:     toPointer(float32(40.0)),
			Inventory: toPointer(10),
			CreatedAt: time.Now().UTC().Round(time.Millisecond),
			UpdatedAt: time.Now().UTC().Round(time.Millisecond),
			Archived:  false,
		}

		newBook, err := store.CreateBook(ctx, b)
		is.NoErr(err)
		compareBooks(is, newBook, b)

		//Archiving the created book.
		archivedBook, err := store.SetBookArchiveStatus(ctx, b.ID, true)
		is.NoErr(err)

		//Changing the status of 'arquived' field of local book to be compare afterwards.
		b.Archived = true

		compareBooks(is, archivedBook, b)
	})

	t.Run("archives an non existing book should return a not found error", func(t *testing.T) {
		is := is.New(t)

		nonexistentBook := book.Book{
			ID:        uuid.New(),
			Name:      "A new book that will not be archived",
			Price:     toPointer(float32(40.0)),
			Inventory: toPointer(10),
			CreatedAt: time.Now().UTC().Round(time.Millisecond),
			UpdatedAt: time.Now().UTC().Round(time.Millisecond),
			Archived:  false,
		}

		archivedBook, err := store.SetBookArchiveStatus(ctx, nonexistentBook.ID, true)
		is.True(errors.Is(err, book.ErrResponseBookNotFound))
		compareBooks(is, archivedBook, book.Book{})
	})

}
func TestUpdateBook(t *testing.T) {
	t.Cleanup(func() {
		teardownDB(t)
	})

	t.Run("updates a book without errors", func(t *testing.T) {
		is := is.New(t)

		// Setting up, creating a book to be fetched.
		b := book.Book{
			ID:        uuid.New(),
			Name:      "A new book to be updated",
			Price:     toPointer(float32(40.0)),
			Inventory: toPointer(10),
			CreatedAt: time.Now().UTC().Round(time.Millisecond),
			UpdatedAt: time.Now().UTC().Round(time.Millisecond),
		}

		newBook, err := store.CreateBook(ctx, b)
		is.NoErr(err)
		compareBooks(is, newBook, b)

		//Updating the created book.
		b.Name = "The book is now updated"
		b.Price = toPointer(float32(50.0))
		b.Inventory = toPointer(9)
		b.UpdatedAt = time.Now().UTC().Round(time.Millisecond)

		updatedBook, err := store.UpdateBook(ctx, b)
		is.NoErr(err)
		compareBooks(is, updatedBook, b)
	})

	t.Run("Updates an non existing book should return a not found error", func(t *testing.T) {
		is := is.New(t)

		nonexistentBook := book.Book{
			ID:        uuid.New(),
			Name:      "A new book that will not be stored",
			Price:     toPointer(float32(40.0)),
			Inventory: toPointer(10),
			CreatedAt: time.Now().UTC().Round(time.Millisecond),
			UpdatedAt: time.Now().UTC().Round(time.Millisecond),
		}

		returnedBook, err := store.UpdateBook(ctx, nonexistentBook)
		is.True(errors.Is(err, book.ErrResponseBookNotFound))
		compareBooks(is, returnedBook, book.Book{})
	})
}

func TestGetBook(t *testing.T) {
	t.Cleanup(func() {
		teardownDB(t)
	})

	t.Run("Gets a book by ID without errors", func(t *testing.T) {
		is := is.New(t)

		// Setting up, creating a book to be fetched.
		b := book.Book{
			ID:        uuid.New(),
			Name:      "A new book`",
			Price:     toPointer(float32(40.0)),
			Inventory: toPointer(10),
			CreatedAt: time.Now().UTC().Round(time.Millisecond),
			UpdatedAt: time.Now().UTC().Round(time.Millisecond),
		}

		newBook, err := store.CreateBook(ctx, b)
		is.NoErr(err)
		compareBooks(is, newBook, b)

		returnedBook, err := store.GetBookByID(ctx, b.ID)
		is.NoErr(err)
		compareBooks(is, returnedBook, b)
	})

	t.Run("Gets an non existing book should return a not found error", func(t *testing.T) {
		is := is.New(t)

		returnedBook, err := store.GetBookByID(ctx, uuid.New())
		is.True(errors.Is(err, book.ErrResponseBookNotFound))
		compareBooks(is, returnedBook, book.Book{})
	})
}

func TestListBooks(t *testing.T) {
	t.Cleanup(func() {
		teardownDB(t)
	})

	is := is.New(t)
	var testBookslist []book.Book
	listSize := 30

	t.Run("List books without errors even if there is no books in the database", func(t *testing.T) {
		is := is.New(t)

		// Write the List Books test here.
		returnedBooks, err := store.ListBooks(ctx, "", 0.00, 9999.99, "name", "asc", true, 30, 0)
		is.NoErr(err)
		is.Equal(returnedBooks, []book.Book{})
	})

	// Setting up, creating books to be listed.
	for i := 0; i < listSize; i++ {
		b := book.Book{
			ID:        uuid.New(),
			Name:      fmt.Sprintf("Book number %06v", i),
			Price:     toPointer(float32((i * 100) + 1)),
			Inventory: toPointer(i + 1),
			CreatedAt: time.Now().UTC().Round(time.Millisecond),
			UpdatedAt: time.Now().UTC().Round(time.Millisecond),
		}

		newBook, err := store.CreateBook(ctx, b)
		is.NoErr(err)
		compareBooks(is, newBook, b)
		testBookslist = append(testBookslist, b)
	}

	t.Run("List all books, no filtering, without errors.", func(t *testing.T) {
		is := is.New(t)

		//Asking all books on the list. Expected 30 books on page 1.
		itemsTotal, err := store.ListBooksTotals(ctx, "", 0.00, 9999.99, true)
		is.NoErr(err)
		is.True(itemsTotal == 30)
		returnedBooks, err := store.ListBooks(ctx, "", 0.00, 9999.99, "name", "asc", true, 1, 30)
		is.NoErr(err)
		for i, expected := range testBookslist {
			compareBooks(is, returnedBooks[i], expected)
		}
	})

	t.Run("List books with limited page size, without errors.", func(t *testing.T) {
		is := is.New(t)

		//Asking 10 books of the list each time.
		for p := 1; p <= 3; p++ {
			itemsTotal, err := store.ListBooksTotals(ctx, "", 0.00, 9999.99, true)
			is.NoErr(err)
			is.True(itemsTotal == 30)
			returnedBooks, err := store.ListBooks(ctx, "", 0.00, 9999.99, "name", "asc", true, p, 10)
			is.NoErr(err)
			is.True(len(returnedBooks) == 10)
			for i, expected := range testBookslist[((p - 1) * 10):(((p - 1) * 10) + 9)] {
				compareBooks(is, returnedBooks[i], expected)
			}
		}
	})

	t.Run("List books without errors filtering by exactly name", func(t *testing.T) {
		is := is.New(t)

		// Testing, by name, each book on the created list.
		for i := 0; i < listSize; i++ {
			returnedBook, err := store.ListBooks(ctx, fmt.Sprintf("Book number %06v", i), 0.00, 9999.99, "name", "asc", true, 1, 30)
			is.NoErr(err)
			is.True(len(returnedBook) == 1)
			compareBooks(is, returnedBook[0], testBookslist[i])
		}
	})

	t.Run("List books without errors filtering by partial name", func(t *testing.T) {
		is := is.New(t)

		// Testing the different part of each name
		for i := 0; i < listSize; i++ {
			returnedBook, err := store.ListBooks(ctx, fmt.Sprintf( /* Book */ "number %06v", i), 0.00, 9999.99, "name", "asc", true, 1, 30)
			is.NoErr(err)
			is.True(len(returnedBook) == 1)
			compareBooks(is, returnedBook[0], testBookslist[i])
		}
		//Testing the common part of all names on the list
		returnedBooks, err := store.ListBooks(ctx, "Book number" /* %06v, i */, 0.00, 9999.99, "name", "asc", true, 1, 30)
		is.NoErr(err)
		is.True(len(returnedBooks) == listSize)
		for i, expected := range testBookslist {
			compareBooks(is, returnedBooks[i], expected)
		}
	})

	t.Run("List books without errors filtering by minimum price", func(t *testing.T) {
		is := is.New(t)

		//Asking all books on the created list with price >= 501
		returnedBooks, err := store.ListBooks(ctx, "", 501.00, 9999.99, "name", "asc", true, 1, 30)
		is.NoErr(err)
		for i, expected := range testBookslist[5:11] {
			compareBooks(is, returnedBooks[i], expected)
		}
	})

	t.Run("List books without errors filtering by maximum price", func(t *testing.T) {
		is := is.New(t)

		//Asking all books on the created list with price <= 501
		returnedBooks, err := store.ListBooks(ctx, "", 00.00, 501.00, "name", "asc", true, 1, 30)
		is.NoErr(err)
		for i, expected := range testBookslist[0:6] {
			compareBooks(is, returnedBooks[i], expected)
		}
	})

	t.Run("List all books without errors ordering by price, ascendent direction", func(t *testing.T) {
		is := is.New(t)

		returnedBooks, err := store.ListBooks(ctx, "", 00.00, 9999.99, "price", "asc", true, 1, 30)
		is.NoErr(err)
		var lastPrice float32 = 0
		for _, v := range returnedBooks {
			is.True(*v.Price >= lastPrice)
			lastPrice = *v.Price
		}
	})

	t.Run("List all books without errors ordering by price, descendent direction", func(t *testing.T) {
		is := is.New(t)

		returnedBooks, err := store.ListBooks(ctx, "", 00.00, 9999.99, "price", "desc", true, 1, 30)
		is.NoErr(err)
		var lastPrice float32 = 9999.99
		for _, v := range returnedBooks {
			is.True(*v.Price <= lastPrice)
			lastPrice = *v.Price
		}
	})

	t.Run("List not archived books without errors", func(t *testing.T) {
		is := is.New(t)
		//Archiving one book of the list
		archivedBook, err := store.SetBookArchiveStatus(ctx, testBookslist[0].ID, true)
		is.NoErr(err)
		is.True(archivedBook.Archived == true)

		// Testing if the returned list has one book less and if all of the returned books are 'false' for 'archived'
		returnedBook, err := store.ListBooks(ctx, "", 0.00, 9999.99, "name", "asc", false, 1, 30)
		is.NoErr(err)
		is.True(len(returnedBook) == (listSize - 1))

		for i := 0; i < (listSize - 1); i++ {
			is.True(returnedBook[i].Archived == false)
		}
	})

	t.Run("Filtering a list by an archived book name returns an empty list, no errors.", func(t *testing.T) {
		is := is.New(t)
		//Book number 000000 was archived on last test.
		returnedBook, err := store.ListBooks(ctx, "Book number 000000", 0.00, 9999.99, "name", "asc", false, 1, 30)
		is.NoErr(err)
		is.True(len(returnedBook) == 0)
	})
}

func TestDownMigrations(t *testing.T) {
	is := is.New(t)
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	is.NoErr(err)

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", "../../../migrations"),
		"postgres", driver)
	is.NoErr(err)

	t.Cleanup(func() {
		is.NoErr(m.Up())
	})

	err = m.Down()
	is.NoErr(err)
	sqlStatement := `SELECT EXISTS (
		SELECT FROM 
			pg_tables
		WHERE 
			schemaname = 'public' AND 
			tablename  = 'bookstable'
		);`
	check := sqlDB.QueryRow(sqlStatement)
	var tableExists bool
	err = check.Scan(&tableExists)
	is.NoErr(err)
	is.True(!tableExists)
}

func TestCreateOrder(t *testing.T) {
	t.Cleanup(func() {
		teardownDB(t)
	})

	t.Run("creates an order with a generic user", func(t *testing.T) {
		is := is.New(t)

		o := book.Order{
			OrderID:     uuid.New(),
			PurchaserID: uuid.New(),
			OrderStatus: "accepting_items",
			CreatedAt:   time.Now().UTC().Round(time.Millisecond),
			UpdatedAt:   time.Now().UTC().Round(time.Millisecond),
		}

		newOrder, err := store.CreateOrder(ctx, o)
		is.NoErr(err)
		compareOrders(is, newOrder, o)
	})
}

func TestListOrderItems(t *testing.T) {
	t.Cleanup(func() {
		teardownDB(t)
	})

	is := is.New(t)
	var testBookslist []book.Book
	listSize := 5

	// Setting up, creating books to be listed.
	for i := 0; i < listSize; i++ {
		b := book.Book{
			ID:        uuid.New(),
			Name:      fmt.Sprintf("Book number %06v", i),
			Price:     toPointer(float32((i * 100) + 1)),
			Inventory: toPointer(i + 1),
			CreatedAt: time.Now().UTC().Round(time.Millisecond),
			UpdatedAt: time.Now().UTC().Round(time.Millisecond),
		}

		newBook, err := store.CreateBook(ctx, b)
		is.NoErr(err)
		compareBooks(is, newBook, b)
		testBookslist = append(testBookslist, b)
	}

	t.Run("lists items from an order without errors", func(t *testing.T) {
		is := is.New(t)

		//creating order to be fetched:
		o := book.Order{
			OrderID:     uuid.New(),
			PurchaserID: uuid.New(),
			OrderStatus: "accepting_items",
			CreatedAt:   time.Now().UTC().Round(time.Millisecond),
			UpdatedAt:   time.Now().UTC().Round(time.Millisecond),
		}
		newOrder, err := store.CreateOrder(ctx, o)
		is.NoErr(err)

		compareOrders(is, newOrder, o)

		//adding the books to an order:
		storedList := []book.OrderItem{}
		bookUnits := 1
		for _, bk := range testBookslist {
			//changing books into itemsAtOrder:
			bkItem := book.OrderItem{
				OrderID:          o.OrderID,
				BookID:           bk.ID,
				BookUnits:        bookUnits,
				BookPriceAtOrder: bk.Price,
				CreatedAt:        time.Now().UTC().Round(time.Millisecond),
				UpdatedAt:        time.Now().UTC().Round(time.Millisecond),
			}

			bookAtOrder, err := store.AddItemToOrder(ctx, bkItem)
			is.NoErr(err)
			storedList = append(storedList, bookAtOrder)
		}

		//testing if it returns a valid list:
		fetchedOrder, fetchedList, err := store.ListOrderItems(ctx, o.OrderID)
		is.NoErr(err)
		compareOrders(is, fetchedOrder, o)
		for i, expected := range storedList {
			compareItemsAtOrder(is, fetchedList[i], expected)
		}
	})

	t.Run("lists items from an inexistent order should return a not found error", func(t *testing.T) {
		is := is.New(t)

		fetchedOrder, fetchedList, err := store.ListOrderItems(ctx, uuid.New())
		is.True(errors.Is(err, book.ErrResponseOrderNotFound))
		compareOrders(is, fetchedOrder, book.Order{})
		is.True(len(fetchedList) == 0)
	})
}

func TestUpdateOrder(t *testing.T) {
	t.Cleanup(func() {
		teardownDB(t)
	})

	is := is.New(t)
	var testBookslist []book.Book
	listSize := 5

	// Setting up, creating books to be added to an order.
	for i := 0; i < listSize; i++ {
		b := book.Book{
			ID:        uuid.New(),
			Name:      fmt.Sprintf("Book number %06v", i),
			Price:     toPointer(float32((i * 100) + 1)),
			Inventory: toPointer(3),
			CreatedAt: time.Now().UTC().Round(time.Millisecond),
			UpdatedAt: time.Now().UTC().Round(time.Millisecond),
		}

		newBook, err := store.CreateBook(ctx, b)
		is.NoErr(err)
		compareBooks(is, newBook, b)
		testBookslist = append(testBookslist, b)
	}

	//creating order to be fetched:
	o := book.Order{
		OrderID:     uuid.New(),
		PurchaserID: uuid.New(),
		OrderStatus: "accepting_items",
		CreatedAt:   time.Now().UTC().Round(time.Millisecond),
		UpdatedAt:   time.Now().UTC().Round(time.Millisecond),
	}
	newOrder, err := store.CreateOrder(ctx, o)
	is.NoErr(err)

	compareOrders(is, newOrder, o)

	t.Run("add an item to an order without errors", func(t *testing.T) {
		is := is.New(t)

		/* PASS IT TO DATABASE LAYER?
		//changing books into itemAtOrder:
		bkItem := book.OrderItem{
			OrderID:          o.OrderID,
			BookID:           testBookslist[0].ID,
			BookUnits:        1,
			BookPriceAtOrder: testBookslist[0].Price,
			CreatedAt:        time.Now().UTC().Round(time.Millisecond),
			UpdatedAt:        time.Now().UTC().Round(time.Millisecond),
		}*/

		updtReq := book.UpdateOrderRequest{
			OrderID:        o.OrderID,
			BookID:         testBookslist[0].ID,
			BookUnitsToAdd: 2,
		}

		bookAtOrder, err := store.UpdateOrder(ctx, updtReq)
		is.NoErr(err)
		is.Equal(bookAtOrder.OrderID, updtReq.OrderID)
		is.Equal(bookAtOrder.BookID, updtReq.BookID)
		is.Equal(bookAtOrder.BookUnits, updtReq.BookUnitsToAdd) //In this test we are ADDING a book to an order, so BookUnits starts from zero.

		//testing if the transaction was really commited:
		fetchedOrder, fetchedList, err := store.ListOrderItems(ctx, o.OrderID)
		is.NoErr(err)
		is.True(fetchedOrder.UpdatedAt.Compare(fetchedOrder.CreatedAt.Round(time.Millisecond)) > 0)
		is.Equal(fetchedList[0].BookID, updtReq.BookID)

		//testing if the book was updated at bookstable:
		bk, err := store.GetBookByID(ctx, updtReq.BookID)
		is.NoErr(err)
		is.True(*bk.Inventory == (*testBookslist[0].Inventory - updtReq.BookUnitsToAdd))
		is.True(bk.UpdatedAt.Compare(bk.CreatedAt.Round(time.Millisecond)) > 0)
	})
}

// compareBooks asserts that two books are equal,
// handling time.Time values correctly.
func compareBooks(is *is.I, a, b book.Book) {
	is.Helper()

	// Make sure we have the correct timestamps.
	is.True(a.CreatedAt.Equal(b.CreatedAt))
	is.True(a.UpdatedAt.Equal(b.UpdatedAt))

	// Overwrite to be able to compare them.
	b.CreatedAt = a.CreatedAt
	b.UpdatedAt = a.UpdatedAt

	// Assert that they are equal.
	is.Equal(a, b)
}

func compareOrders(is *is.I, a, b book.Order) {
	is.Helper()

	// Make sure we have the correct timestamps.
	is.True(a.CreatedAt.Equal(b.CreatedAt))
	is.True(a.UpdatedAt.Equal(b.UpdatedAt))

	// Overwrite to be able to compare them.
	b.CreatedAt = a.CreatedAt
	b.UpdatedAt = a.UpdatedAt

	// Assert that they are equal.
	is.Equal(a, b)
}

func compareItemsAtOrder(is *is.I, a, b book.OrderItem) {
	is.Helper()

	// Make sure we have the correct timestamps.
	is.True(a.CreatedAt.Equal(b.CreatedAt))
	is.True(a.UpdatedAt.Equal(b.UpdatedAt))

	// Overwrite to be able to compare them.
	b.CreatedAt = a.CreatedAt
	b.UpdatedAt = a.UpdatedAt

	// Assert that they are equal.
	is.Equal(a, b)
}

func toPointer[T any](v T) *T {
	return &v
}

func teardownDB(t *testing.T) {
	is := is.New(t)

	// Truncating books table, cleaning up all the records.
	result, err := sqlDB.Exec(`TRUNCATE TABLE public.bookstable CASCADE`) //SHOULD WE TRUNCATE THE OTHER TABLES TOO?????
	is.NoErr(err)

	_, err = result.RowsAffected()
	is.NoErr(err)
}
