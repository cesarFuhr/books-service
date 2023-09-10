package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

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
			CreatedAt: time.Now().UTC().Round(time.Millisecond),
			UpdatedAt: time.Now().UTC().Round(time.Millisecond),
		}

		newBook, err := storeOnDB(b)
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
		b := Book{
			ID:        uuid.New(),
			Name:      "A new book to be archived",
			Price:     toPointer(float32(40.0)),
			Inventory: toPointer(10),
			CreatedAt: time.Now().UTC().Round(time.Millisecond),
			UpdatedAt: time.Now().UTC().Round(time.Millisecond),
			Archived:  false,
		}

		newBook, err := storeOnDB(b)
		is.NoErr(err)
		compareBooks(is, newBook, b)

		//Archiving the created book.
		archivedBook, err := archiveStatusOnDB(b.ID, true)
		is.NoErr(err)

		//Changing the status of 'arquived' field of local book to be compare afterwards.
		b.Archived = true

		compareBooks(is, archivedBook, b)
	})

	t.Run("archives an non existing book should return a not found error", func(t *testing.T) {
		is := is.New(t)

		nonexistentBook := Book{
			ID:        uuid.New(),
			Name:      "A new book that will not be archived",
			Price:     toPointer(float32(40.0)),
			Inventory: toPointer(10),
			CreatedAt: time.Now().UTC().Round(time.Millisecond),
			UpdatedAt: time.Now().UTC().Round(time.Millisecond),
			Archived:  false,
		}

		archivedBook, err := archiveStatusOnDB(nonexistentBook.ID, true)
		is.True(errors.Is(err, errResponseBookNotFound))
		compareBooks(is, archivedBook, Book{})
	})

}
func TestUpdateBook(t *testing.T) {
	t.Cleanup(func() {
		teardownDB(t)
	})

	t.Run("updates a book without errors", func(t *testing.T) {
		is := is.New(t)

		// Setting up, creating a book to be fetched.
		b := Book{
			ID:        uuid.New(),
			Name:      "A new book to be updated",
			Price:     toPointer(float32(40.0)),
			Inventory: toPointer(10),
			CreatedAt: time.Now().UTC().Round(time.Millisecond),
			UpdatedAt: time.Now().UTC().Round(time.Millisecond),
		}

		newBook, err := storeOnDB(b)
		is.NoErr(err)
		compareBooks(is, newBook, b)

		//Updating the created book.
		b.Name = "The book is now updated"
		b.Price = toPointer(float32(50.0))
		b.Inventory = toPointer(9)
		b.UpdatedAt = time.Now().UTC().Round(time.Millisecond)

		updatedBook, err := updateOnDB(b)
		is.NoErr(err)
		compareBooks(is, updatedBook, b)
	})

	t.Run("Updates an non existing book should return a not found error", func(t *testing.T) {
		is := is.New(t)

		nonexistentBook := Book{
			ID:        uuid.New(),
			Name:      "A new book that will not be stored",
			Price:     toPointer(float32(40.0)),
			Inventory: toPointer(10),
			CreatedAt: time.Now().UTC().Round(time.Millisecond),
			UpdatedAt: time.Now().UTC().Round(time.Millisecond),
		}

		returnedBook, err := updateOnDB(nonexistentBook)
		is.True(errors.Is(err, errResponseBookNotFound))
		compareBooks(is, returnedBook, Book{})
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
			CreatedAt: time.Now().UTC().Round(time.Millisecond),
			UpdatedAt: time.Now().UTC().Round(time.Millisecond),
		}

		newBook, err := storeOnDB(b)
		is.NoErr(err)
		compareBooks(is, newBook, b)

		// Write the Get Book test here.
		returnedBook, err := searchById(b.ID)
		is.NoErr(err)
		compareBooks(is, returnedBook, b)
	})

	t.Run("Gets an non existing book should return a not found error", func(t *testing.T) {
		is := is.New(t)

		// Write the Get Book test here.
		returnedBook, err := searchById(uuid.New())
		is.True(errors.Is(err, errResponseBookNotFound))
		compareBooks(is, returnedBook, Book{})
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
		returnedBooks, err := listBooks("", 0.00, 9999.99, "name", "asc", true)
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
			CreatedAt: time.Now().UTC().Round(time.Millisecond),
			UpdatedAt: time.Now().UTC().Round(time.Millisecond),
		}

		newBook, err := storeOnDB(b)
		is.NoErr(err)
		compareBooks(is, newBook, b)
		testBookslist = append(testBookslist, b)
	}

	t.Run("List all books, no filtering, without errors", func(t *testing.T) {
		is := is.New(t)

		//Asking all books on the list
		returnedBooks, err := listBooks("", 0.00, 9999.99, "name", "asc", true)
		is.NoErr(err)
		for i, expected := range testBookslist {
			compareBooks(is, returnedBooks[i], expected)
		}
	})

	t.Run("List books without errors filtering by exactly name", func(t *testing.T) {
		is := is.New(t)

		// Testing, by name, each book on the created list.
		for i := 0; i < listSize; i++ {
			returnedBook, err := listBooks(fmt.Sprintf("Book number %06v", i), 0.00, 9999.99, "name", "asc", true)
			is.NoErr(err)
			is.True(len(returnedBook) == 1)
			compareBooks(is, returnedBook[0], testBookslist[i])
		}
	})

	t.Run("List books without errors filtering by partial name", func(t *testing.T) {
		is := is.New(t)

		// Testing the different part of each name
		for i := 0; i < listSize; i++ {
			returnedBook, err := listBooks(fmt.Sprintf( /* Book */ "number %06v", i), 0.00, 9999.99, "name", "asc", true)
			is.NoErr(err)
			is.True(len(returnedBook) == 1)
			compareBooks(is, returnedBook[0], testBookslist[i])
		}
		//Testing the common part of all names on the list
		returnedBooks, err := listBooks("Book number" /* %06v, i */, 0.00, 9999.99, "name", "asc", true)
		is.NoErr(err)
		is.True(len(returnedBooks) == listSize)
		for i, expected := range testBookslist {
			compareBooks(is, returnedBooks[i], expected)
		}
	})

	t.Run("List books without errors filtering by minimum price", func(t *testing.T) {
		is := is.New(t)

		//Asking all books on the created list with price >= 501
		returnedBooks, err := listBooks("", 501.00, 9999.99, "name", "asc", true)
		is.NoErr(err)
		for i, expected := range testBookslist[5:11] {
			compareBooks(is, returnedBooks[i], expected)
		}
	})

	t.Run("List books without errors filtering by maximum price", func(t *testing.T) {
		is := is.New(t)

		//Asking all books on the created list with price <= 501
		returnedBooks, err := listBooks("", 00.00, 501.00, "name", "asc", true)
		is.NoErr(err)
		for i, expected := range testBookslist[0:6] {
			compareBooks(is, returnedBooks[i], expected)
		}
	})

	t.Run("List all books without errors ordering by price, ascendent direction", func(t *testing.T) {
		is := is.New(t)

		returnedBooks, err := listBooks("", 00.00, 9999.99, "price", "asc", true)
		is.NoErr(err)
		var lastPrice float32 = 0
		for _, v := range returnedBooks {
			is.True(*v.Price >= lastPrice)
			lastPrice = *v.Price
		}
	})

	t.Run("List all books without errors ordering by price, descendent direction", func(t *testing.T) {
		is := is.New(t)

		returnedBooks, err := listBooks("", 00.00, 9999.99, "price", "desc", true)
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
		archivedBook, err := archiveStatusOnDB(testBookslist[0].ID, true)
		is.NoErr(err)
		is.True(archivedBook.Archived == true)

		// Testing if the returned list has one book less and if all of the returned books are 'false' for 'archived'
		returnedBook, err := listBooks("", 0.00, 9999.99, "name", "asc", false)
		is.NoErr(err)
		is.True(len(returnedBook) == (listSize - 1))

		for i := 0; i < (listSize - 1); i++ {
			is.True(returnedBook[i].Archived == false)
		}
	})

	t.Run("Filtering a list by an archived book name returns an empty list, no errors.", func(t *testing.T) {
		is := is.New(t)
		//Book number 000000 was archived on last test.
		returnedBook, err := listBooks("Book number 000000", 0.00, 9999.99, "name", "asc", false)
		is.NoErr(err)
		is.True(len(returnedBook) == 0)
	})
}

func TestDownMigrations(t *testing.T) {
	is := is.New(t)
	err := mGlobal.Down()
	is.NoErr(err)
	sqlStatement := `SELECT EXISTS (
		SELECT FROM 
			pg_tables
		WHERE 
			schemaname = 'public' AND 
			tablename  = 'bookstable'
		);`
	check := dbObjectGlobal.QueryRow(sqlStatement)
	var tableExists bool
	err = check.Scan(&tableExists)
	is.NoErr(err)
	is.True(tableExists == false)
}

// compareBooks asserts that two books are equal,
// handling time.Time values correctly.
func compareBooks(is *is.I, a, b Book) {
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
	result, err := dbObjectGlobal.Exec(`TRUNCATE TABLE public.bookstable`)
	is.NoErr(err)

	_, err = result.RowsAffected()
	is.NoErr(err)
}
