package book

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/google/uuid"
)

type ServiceAPI interface {
	ArchiveBook(ctx context.Context, id uuid.UUID) (Book, error)
	CreateBook(ctx context.Context, req CreateBookRequest) (Book, error)
	GetBook(ctx context.Context, id uuid.UUID) (Book, error)
	ListBooks(ctx context.Context, params ListBooksRequest) (PagedBooks, error)
	CreateOrder(ctx context.Context, user_id uuid.UUID) (Order, error)
	UpdateBook(ctx context.Context, req UpdateBookRequest) (Book, error)
	UpdateOrderTx(ctx context.Context, updtReq UpdateOrderRequest) (Order, error)
	ListOrderItems(ctx context.Context, order_id uuid.UUID) (Order, error)
}

type Repository interface {
	SetBookArchiveStatus(ctx context.Context, id uuid.UUID, archived bool) (Book, error)
	CreateBook(ctx context.Context, bookEntry Book) (Book, error)
	GetBookByID(ctx context.Context, id uuid.UUID) (Book, error)
	ListBooks(ctx context.Context, name string, minPrice32, maxPrice32 float32, sortBy, sortDirection string, archived bool, page, pageSize int) ([]Book, error)
	ListBooksTotals(ctx context.Context, name string, minPrice32, maxPrice32 float32, archived bool) (int, error)
	UpdateBook(ctx context.Context, bookEntry Book) (Book, error)
	CreateOrder(ctx context.Context, newOrder Order) (Order, error)
	ListOrderItems(ctx context.Context, order_id uuid.UUID) (Order, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Repository, driver.Tx, error)
	GetOrderItem(ctx context.Context, orderID uuid.UUID, bookID uuid.UUID) (OrderItem, error)
	UpdateOrderRow(ctx context.Context, orderID uuid.UUID) error
	UpsertOrderItem(ctx context.Context, orderID uuid.UUID, itemToUpdt OrderItem) (OrderItem, error)
	DeleteOrderItem(ctx context.Context, orderID uuid.UUID, bookID uuid.UUID) error
}

type Notifier interface {
	BookCreated(ctx context.Context, createdBook Book) error
}

type Service struct {
	repo                 Repository
	ntf                  Notifier
	notificationsTimeout time.Duration
}

func NewService(repo Repository, ntf Notifier, notificationsTimeout time.Duration) *Service {
	return &Service{
		repo:                 repo,
		ntf:                  ntf,
		notificationsTimeout: notificationsTimeout,
	}
}

func (s *Service) ArchiveBook(ctx context.Context, id uuid.UUID) (Book, error) {
	archived := true
	return s.repo.SetBookArchiveStatus(ctx, id, archived)
}

type CreateBookRequest struct {
	Name      string
	Price     *float32
	Inventory *int
}

func (s *Service) CreateBook(ctx context.Context, req CreateBookRequest) (Book, error) {
	createdAt := time.Now().UTC().Round(time.Millisecond) //Atribute creating and updating time to the new entry. UpdateAt can change later.
	newBook := Book{
		ID:        uuid.New(), //Atribute an ID to the entry
		Name:      req.Name,
		Price:     req.Price,
		Inventory: req.Inventory,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
		//Archived is set to false by defalut inside database
	}

	b, err := s.repo.CreateBook(ctx, newBook)
	if err == nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), s.notificationsTimeout)
			defer cancel()
			err := s.ntf.BookCreated(ctx, newBook)
			if err != nil {
				log.Println(err)
			}
		}()
	}
	return b, err
}

type UpdateBookRequest struct {
	ID        uuid.UUID
	Name      string
	Price     *float32
	Inventory *int
}

func (s *Service) UpdateBook(ctx context.Context, req UpdateBookRequest) (Book, error) {
	updatedAt := time.Now().UTC().Round(time.Millisecond) //Atribute a new updating time to the new entry.
	updateBook := Book{
		ID:        req.ID,
		Name:      req.Name,
		Price:     req.Price,
		Inventory: req.Inventory,
		//CreatedAt will not change
		UpdatedAt: updatedAt,
		//Archived will not change
	}
	return s.repo.UpdateBook(ctx, updateBook)
}

func (s *Service) GetBook(ctx context.Context, id uuid.UUID) (Book, error) {
	return s.repo.GetBookByID(ctx, id)
}

type PagedBooks struct {
	PageCurrent int
	PageTotal   int
	PageSize    int
	ItemsTotal  int
	Results     []Book
}

type ListBooksRequest struct {
	Name          string
	MinPrice      float32
	MaxPrice      float32
	SortBy        string
	SortDirection string
	Archived      bool
	Page          int
	PageSize      int
}

func (s *Service) ListBooks(ctx context.Context, params ListBooksRequest) (PagedBooks, error) {
	itemsTotal, err := s.repo.ListBooksTotals(ctx, params.Name, params.MinPrice, params.MaxPrice, params.Archived)
	if err != nil {
		return PagedBooks{}, fmt.Errorf("error on call to ListBookTotals: %w ", err)
	}

	if itemsTotal == 0 {
		noBooks := PagedBooks{
			PageCurrent: 0,
			PageTotal:   0,
			PageSize:    0,
			ItemsTotal:  0,
			Results:     []Book{},
		}
		return noBooks, nil
	}

	pagesTotal, err := pagination(params.Page, params.PageSize, itemsTotal)
	if err != nil {
		return PagedBooks{}, err
	}

	//Ask filtered list to db:
	returnedBooks, err := s.repo.ListBooks(ctx, params.Name, params.MinPrice, params.MaxPrice, params.SortBy, params.SortDirection, params.Archived, params.Page, params.PageSize)
	if err != nil {
		return PagedBooks{}, fmt.Errorf("error on call to ListBooks: %w", err)
	}

	pageOfBooksList := PagedBooks{
		PageCurrent: params.Page,
		PageTotal:   pagesTotal,
		PageSize:    params.PageSize,
		ItemsTotal:  itemsTotal,
		Results:     returnedBooks,
	}

	return pageOfBooksList, nil
}

/*Calculates the pagination.*/
func pagination(page, pageSize, itemsTotal int) (pagesTotal int, err error) {
	pagesTotal = int(math.Ceil(float64(itemsTotal) / float64(pageSize)))
	if page > pagesTotal {
		return 0, ErrResponseQueryPageOutOfRange
	}

	return pagesTotal, nil
}
