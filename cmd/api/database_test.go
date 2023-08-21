package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/matryer/is"
)

// TestMain is called before all the tests run.
// Usually is where we add logic to initialise resources.
func TestMain(m *testing.M) {
	// Setting up the database for tests.
	os.Setenv("DATABASE_URL", "postgres://root:root@localhost:5432/booksdb?sslmode=disable")
	db, err := connectDb()
	if err != nil {
		log.Fatalln(err)
	}

	dbObjectGlobal = db

	os.Setenv("DATABASE_MIGRATIONS_PATH", "../../migrations")
	err = migrationUp()
	if err != nil {
		log.Fatalln(err)
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

		b := Book{
			ID:        uuid.New(),
			Name:      "A new book`",
			Price:     toPointer(float32(40.0)),
			Inventory: toPointer(10),
		}

		newBook, err := storeOnDB(b)
		is.NoErr(err)
		is.Equal(newBook, b)
	})
}

func TestGetBook(t *testing.T) {
	t.Cleanup(func() {
		teardownDB(t)
	})

	t.Run("Gets a book by ID without errors", func(t *testing.T) {
		is := is.New(t)

		// Setting up, creating a book to be fetched.
		b := Book{
			ID:        uuid.New(),
			Name:      "A new book`",
			Price:     toPointer(float32(40.0)),
			Inventory: toPointer(10),
		}

		newBook, err := storeOnDB(b)
		is.NoErr(err)
		is.Equal(newBook, b)

		// Write the Get Book test here.
		returnedBook, err := searchById(b.ID)
		is.NoErr(err)
		is.Equal(returnedBook, b)
	})

	t.Run("Gets an non existing book should return a not found error", func(t *testing.T) {
		is := is.New(t)

		// Write the Get Book test here.
		returnedBook, err := searchById(uuid.New())
		is.True(errors.Is(err, errBookNotFound))
		is.Equal(returnedBook, Book{})
	})
}

func TestListBooks(t *testing.T) {
	t.Cleanup(func() {
		teardownDB(t)
	})

	is := is.New(t)
	var testBookslist []Book
	listSize := 11

	t.Run("List books without errors even if there is no books in the database", func(t *testing.T) {
		is := is.New(t)

		// Write the List Books test here.
		returnedBooks, err := listBooks("", 0.00, 9999.99)
		is.NoErr(err)
		is.Equal(returnedBooks, []Book{})
	})

	// Setting up, creating books to be listed.
	for i := 0; i < listSize; i++ {
		b := Book{
			ID:        uuid.New(),
			Name:      fmt.Sprintf("Book number %06v", i),
			Price:     toPointer(float32((i * 100) + 1)),
			Inventory: toPointer(i + 1),
		}

		newBook, err := storeOnDB(b)
		is.NoErr(err)
		is.Equal(newBook, b)
		testBookslist = append(testBookslist, b)
	}

	t.Run("List all books, no filtering, without errors", func(t *testing.T) {
		is := is.New(t)

		//Asking all books on the list
		returnedBooks, err := listBooks("", 0.00, 9999.99)
		is.NoErr(err)
		is.Equal(returnedBooks, testBookslist)
	})

	t.Run("List books without errors filtering by name", func(t *testing.T) {
		is := is.New(t)

		// Testing, by name, each book on the created list.
		for i := 0; i < listSize; i++ {
			returnedBook, err := listBooks(fmt.Sprintf("Book number %06v", i), 0.00, 9999.99)
			is.NoErr(err)
			is.True(len(returnedBook) == 1)
			is.Equal(returnedBook[0], testBookslist[i])
		}
	})

	t.Run("List books without errors filtering by minimum price", func(t *testing.T) {
		is := is.New(t)

		//Asking all books on the created list with price >= 501
		returnedBooks, err := listBooks("", 501.00, 9999.99)
		is.NoErr(err)
		is.Equal(returnedBooks, testBookslist[5:11])
	})

	t.Run("List books without errors filtering by maximum price", func(t *testing.T) {
		is := is.New(t)

		//Asking all books on the created list with price <= 501
		returnedBooks, err := listBooks("", 00.00, 501.00)
		is.NoErr(err)
		is.Equal(returnedBooks, testBookslist[0:6])
	})
}

func toPointer[T any](v T) *T {
	return &v
}

func teardownDB(t *testing.T) {
	is := is.New(t)

	// Truncating books table, cleaning up all the records.
	result, err := dbObjectGlobal.Exec(`TRUNCATE TABLE public.bookstable`)
	is.NoErr(err)

	_, err = result.RowsAffected()
	is.NoErr(err)
}