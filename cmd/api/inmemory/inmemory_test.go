package inmemory_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/books-service/cmd/api/book"
	"github.com/books-service/cmd/api/inmemory"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

var ctx context.Context = context.Background()

func TestCreateBook(t *testing.T) {
	store, err := inmemory.NewInMemoryStore()
	if err != nil {
		log.Fatalln(err)
	}

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
	store, err := inmemory.NewInMemoryStore()
	if err != nil {
		log.Fatalln(err)
	}

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

		archivedBook, err := store.SetBookArchiveStatus(ctx, uuid.New(), true)
		is.True(errors.Is(err, book.ErrResponseBookNotFound))
		compareBooks(is, archivedBook, book.Book{})
	})
}

func TestUpdateBook(t *testing.T) {
	store, err := inmemory.NewInMemoryStore()
	if err != nil {
		log.Fatalln(err)
	}

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
	store, err := inmemory.NewInMemoryStore()
	if err != nil {
		log.Fatalln(err)
	}

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

	store, err := inmemory.NewInMemoryStore()
	if err != nil {
		log.Fatalln(err)
	}

	is := is.New(t)
	var testBookslist []book.Book
	listSize := 30

	t.Run("List books without errors even if there is no books in the database", func(t *testing.T) {
		is := is.New(t)

		returnedBooks, err := store.ListBooks(ctx, "", 0.00, 9999.99, "name", "asc", true, 1, 10)
		is.NoErr(err)
		is.Equal(returnedBooks, []book.Book{})

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
	})

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

func TestCreateOrder(t *testing.T) {
	store, err := inmemory.NewInMemoryStore()
	if err != nil {
		log.Fatalln(err)
	}

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
	store, err := inmemory.NewInMemoryStore()
	if err != nil {
		log.Fatalln(err)
	}

	is := is.New(t)
	var testBookslist []book.Book
	listSize := 5

	// Setting up, creating books to be listed.
	for i := 0; i < listSize; i++ {
		b := book.Book{
			ID:        uuid.New(),
			Name:      fmt.Sprintf("Book number %06v", i),
			Price:     toPointer(float32(2)),
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
				BookID:           bk.ID,
				BookName:         bk.Name,
				BookUnits:        bookUnits,
				BookPriceAtOrder: bk.Price,
			}

			bookAtOrder, err := store.UpsertOrderItem(ctx, o.OrderID, bkItem)
			time.Sleep(1 * time.Millisecond)
			is.NoErr(err)
			storedList = append(storedList, bookAtOrder)
		}
		o.Items = storedList
		o.TotalPrice = 10 //Each item at list has 1 unit with price 2. List size is 5. So total price should result 10.

		//testing if it returns a valid list:
		fetchedOrder, err := store.ListOrderItems(ctx, o.OrderID)
		is.NoErr(err)
		compareOrders(is, fetchedOrder, o)
	})

	t.Run("lists items from an inexistent order should return a not found error", func(t *testing.T) {
		is := is.New(t)

		fetchedOrder, err := store.ListOrderItems(ctx, uuid.New())
		is.True(errors.Is(err, book.ErrResponseOrderNotFound))
		compareOrders(is, fetchedOrder, book.Order{})
		is.True(len(fetchedOrder.Items) == 0)
	})
}

// Tests all methods of the transaction togheter
func TestUpdateOrderTx(t *testing.T) {
	store, err := inmemory.NewInMemoryStore()
	if err != nil {
		log.Fatalln(err)
	}

	is := is.New(t)
	var testBookslist []book.Book
	listSize := 5
	createdNow := time.Now().UTC().Round(time.Millisecond)
	// Setting up, creating books to be added to an order.
	for i := 0; i < listSize; i++ {

		b := book.Book{
			ID:        uuid.New(),
			Name:      fmt.Sprintf("Book number %06v", i),
			Price:     toPointer(float32((i * 100) + 1)),
			Inventory: toPointer(10),
			CreatedAt: createdNow,
			UpdatedAt: createdNow,
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
		CreatedAt:   createdNow,
		UpdatedAt:   createdNow,
	}
	newOrder, err := store.CreateOrder(ctx, o)
	is.NoErr(err)

	compareOrders(is, newOrder, o)

	t.Run("add an item to an order without errors", func(t *testing.T) {
		is := is.New(t)

		//Setting up variables to this subtest:
		OrderID := o.OrderID
		BookID := testBookslist[0].ID
		BookUnitsToAdd := 2 //In this subtest we are ADDING a book to an order, so BookUnits starts from zero and is supposed to result 2.

		txRepo, tx, err := store.BeginTx(ctx, nil)
		is.NoErr(err)

		defer func() {
			rollbackErr := tx.Rollback()
			is.True(errors.Is(rollbackErr, nil))
		}()

		time.Sleep(time.Millisecond)
		err = txRepo.UpdateOrderRow(ctx, OrderID) //changes field 'updated_at' and checks if the order is 'accepting_items'
		is.NoErr(err)

		//Testing if there are sufficient inventory of the book asked, and if is not archived:
		bk, err := txRepo.GetBookByID(ctx, BookID)
		is.NoErr(err)
		is.True(!bk.Archived)
		balance := *bk.Inventory - BookUnitsToAdd
		is.True(balance >= 0)

		//Testing if the book is already at the order. It is expected not to be.
		bookAtOrder, err := txRepo.GetOrderItem(ctx, OrderID, BookID)
		is.Equal(bookAtOrder, book.OrderItem{})
		is.True(errors.Is(err, book.ErrResponseBookNotAtOrder))

		//In this test case, the book is not at the order yet, and BookUnitsToAdd is 2, so the orderItem must be created:
		bkItem := book.OrderItem{
			BookID:           BookID,
			BookName:         "New Book to test",
			BookUnits:        BookUnitsToAdd,
			BookPriceAtOrder: bk.Price,
			//Created_at and Updated_at fields will be set properly at database layer
		}

		bookAtOrder, err = txRepo.UpsertOrderItem(ctx, OrderID, bkItem)
		time.Sleep(time.Millisecond) //To mock comparison between CreatedAt an UpdatedAt in the next substest
		is.NoErr(err)
		is.Equal(bkItem.BookID, bookAtOrder.BookID)
		//is.Equal(bkItem.BookName, bookAtOrder.BookName)
		is.Equal(bkItem.BookUnits, bookAtOrder.BookUnits) //Expected to be equal because the book was just created.
		is.Equal(bkItem.BookPriceAtOrder, bookAtOrder.BookPriceAtOrder)
		is.True(bookAtOrder.UpdatedAt.Compare(bookAtOrder.CreatedAt.Round(time.Millisecond)) == 0) //Expected to be equal because the book was just created.

		//Updating book inventory acordingly at bookstable:
		*bk.Inventory = balance
		bk.UpdatedAt = time.Now().UTC().Round(time.Millisecond).Add(time.Millisecond)
		bkUpdt, err := txRepo.UpdateBook(ctx, bk)
		is.NoErr(err)
		compareBooks(is, bk, bkUpdt)

		err = tx.Commit()
		is.NoErr(err)

		//testing if the order table was correctly updated:
		fetchedOrder, err := store.ListOrderItems(ctx, OrderID)
		is.NoErr(err)
		is.Equal(bookAtOrder, fetchedOrder.Items[0])
		is.True(fetchedOrder.UpdatedAt.Compare(fetchedOrder.CreatedAt.Round(time.Millisecond)) > 0)
		is.True(fetchedOrder.Items[0].UpdatedAt.Compare(fetchedOrder.Items[0].CreatedAt.Round(time.Millisecond)) == 0)

		//testing if the book was updated at bookstable:
		fetchedBook, err := store.GetBookByID(ctx, BookID)
		is.NoErr(err)
		is.True(*fetchedBook.Inventory == 8) //10 - 2 = 8
		is.True(fetchedBook.UpdatedAt.Compare(fetchedBook.CreatedAt.Round(time.Millisecond)) > 0)
	})

	t.Run("update an item at the order without errors", func(t *testing.T) {
		is := is.New(t)

		//Setting up variables to this subtest:
		OrderID := o.OrderID
		BookID := testBookslist[0].ID
		BookUnitsToAdd := 3 //In the last subtest we already added 2 book units to this order, so this update must result 5.

		txRepo, tx, err := store.BeginTx(ctx, nil) //creates a new 'Store' with same sql.db, but with a sql.tx as the 'Executor'
		is.NoErr(err)

		defer func() {
			rollbackErr := tx.Rollback()
			is.True(errors.Is(rollbackErr, nil))
		}()

		err = txRepo.UpdateOrderRow(ctx, OrderID) //changes field 'updated_at' and checks if the order is 'accepting_items'
		is.NoErr(err)

		//Testing if there are sufficient inventory of the book asked, and if is not archived:
		bk, err := txRepo.GetBookByID(ctx, BookID)
		is.NoErr(err)
		is.True(!bk.Archived)
		balance := *bk.Inventory - BookUnitsToAdd
		is.True(balance >= 0)

		//The book is expected to be at the order, now. Some of its fields will be overwritten:
		bkItem := book.OrderItem{
			BookID:           BookID,
			BookName:         "New Book to test",
			BookUnits:        5, //In the last subtest we already added 2 book units to this order, so this update must result 5.
			BookPriceAtOrder: bk.Price,
			//Created_at and Updated_at fields will be set properly at database layer
		}

		bookAtOrder, err := txRepo.UpsertOrderItem(ctx, OrderID, bkItem)
		is.NoErr(err)
		is.Equal(bookAtOrder.BookID, BookID)
		is.Equal(bookAtOrder.BookUnits, 5)                                                        //In the last subtest we already added 2 book units to this order, so this update must result 5.
		is.True(bookAtOrder.UpdatedAt.Compare(bookAtOrder.CreatedAt.Round(time.Millisecond)) > 0) //Assuring it was UPDATED at order

		//Updating book inventory acordingly at bookstable:
		*bk.Inventory = balance
		bk.UpdatedAt = time.Now().UTC().Round(time.Millisecond)
		bkUpdt, err := txRepo.UpdateBook(ctx, bk)
		is.NoErr(err)
		compareBooks(is, bk, bkUpdt)

		err = tx.Commit()
		is.NoErr(err)

		//testing if the book was correctly updated at book_orders table:
		fetchedOrder, err := store.ListOrderItems(ctx, OrderID)
		is.NoErr(err)
		is.Equal(bookAtOrder, fetchedOrder.Items[0])
		is.True(fetchedOrder.UpdatedAt.Compare(fetchedOrder.CreatedAt.Round(time.Millisecond)) > 0)
		is.True(fetchedOrder.Items[0].UpdatedAt.Compare(fetchedOrder.Items[0].CreatedAt.Round(time.Millisecond)) > 0)

		//testing if the book was updated at bookstable:
		fetchedBook, err := store.GetBookByID(ctx, BookID)
		is.NoErr(err)
		is.True(*fetchedBook.Inventory == 5) //8 - 3 = 5
		is.True(fetchedBook.UpdatedAt.Compare(fetchedBook.CreatedAt.Round(time.Millisecond)) > 0)

	})

	t.Run("update an item at the order subtracting all book units, restoring them to inventory, without errors", func(t *testing.T) {
		is := is.New(t)

		//Setting up variables to this subtest:
		OrderID := o.OrderID
		BookID := testBookslist[0].ID
		BookUnitsToAdd := -5 //In the last subtests we already added 2 + 3 book units to this order, so this update must result 0.

		txRepo, tx, err := store.BeginTx(ctx, nil) //creates a new 'Store' with same sql.db, but with a sql.tx as the 'Executor'
		is.NoErr(err)

		defer func() {
			rollbackErr := tx.Rollback()
			is.True(errors.Is(rollbackErr, nil))
		}()

		err = txRepo.UpdateOrderRow(ctx, OrderID) //changes field 'updated_at' and checks if the order is 'accepting_items'
		is.NoErr(err)

		//Testing if there are sufficient inventory of the book asked, and if is not archived:
		bk, err := txRepo.GetBookByID(ctx, BookID)
		is.NoErr(err)
		is.True(!bk.Archived)
		balance := *bk.Inventory - BookUnitsToAdd
		is.True(balance >= 0)

		//Testing if the book is already at the order. It is expected to be, now.
		bookAtOrder, err := txRepo.GetOrderItem(ctx, OrderID, BookID)
		is.NoErr(err)
		is.Equal(bookAtOrder.BookID, BookID)
		is.Equal(bookAtOrder.BookUnits, 5) //In the last subtests we already added 2 + 3 book units to this order, so now it must be 5. After update it will result 0.

		//As book_units becomes zero, the row must be excluded from book_orders table. Even so, it must be updated at bookstable.
		err = txRepo.DeleteOrderItem(ctx, OrderID, BookID)
		is.NoErr(err)

		//Updating book inventory acordingly at bookstable:
		*bk.Inventory = balance
		bk.UpdatedAt = time.Now().UTC().Round(time.Millisecond)
		bkUpdt, err := txRepo.UpdateBook(ctx, bk)
		is.NoErr(err)
		compareBooks(is, bk, bkUpdt)

		err = tx.Commit()
		is.NoErr(err)

		//testing if the book was correctly deleted from book_orders table:
		fetchedOrder, err := store.ListOrderItems(ctx, o.OrderID)
		is.NoErr(err)
		is.True(fetchedOrder.UpdatedAt.Compare(fetchedOrder.CreatedAt.Round(time.Millisecond)) > 0)
		is.True(len(fetchedOrder.Items) == 0) //The list must be empty

		//testing if the book was updated at bookstable:
		fetchedBook, err := store.GetBookByID(ctx, BookID)
		is.NoErr(err)
		is.True(*fetchedBook.Inventory == 10) //5 - (-5) = 10

	})
}

func toPointer[T any](v T) *T {
	return &v
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
