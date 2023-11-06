package book

import (
	"math"
	"time"

	"github.com/google/uuid"
)

type ServiceAPI interface {
	ArchiveBook(id uuid.UUID) (Book, error)
	CreateBook(bookEntry EntryBookRequest) (Book, error)
	GetBook(id uuid.UUID) (Book, error)
	ListBooks(params ListBooksRequest) (PagedBooks, error)
	UpdateBook(bookEntry EntryBookRequest, id uuid.UUID) (Book, error)
}

type Repository interface {
	SetBookArchiveStatus(id uuid.UUID, archived bool) (Book, error)
	CreateBook(bookEntry Book) (Book, error)
	GetBookByID(id uuid.UUID) (Book, error)
	ListBooks(name string, minPrice32, maxPrice32 float32, sortBy, sortDirection string, archived bool, page, pageSize int) ([]Book, error)
	ListBooksTotals(name string, minPrice32, maxPrice32 float32, archived bool) (int, error)
	UpdateBook(bookEntry Book) (Book, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ArchiveBook(id uuid.UUID) (Book, error) {
	archived := true
	return s.repo.SetBookArchiveStatus(id, archived)
}

type EntryBookRequest struct {
	Name      string
	Price     *float32
	Inventory *int
}

func (s *Service) CreateBook(bookEntry EntryBookRequest) (Book, error) {
	createdAt := time.Now().UTC().Round(time.Millisecond) //Atribute creating and updating time to the new entry. UpdateAt can change later.
	newBook := Book{
		ID:        uuid.New(), //Atribute an ID to the entry
		Name:      bookEntry.Name,
		Price:     bookEntry.Price,
		Inventory: bookEntry.Inventory,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
		//Archived is set to false by defalut inside database
	}
	return s.repo.CreateBook(newBook)
}

func (s *Service) UpdateBook(bookEntry EntryBookRequest, id uuid.UUID) (Book, error) {
	updatedAt := time.Now().UTC().Round(time.Millisecond) //Atribute a new updating time to the new entry.
	updateBook := Book{
		ID:        id,
		Name:      bookEntry.Name,
		Price:     bookEntry.Price,
		Inventory: bookEntry.Inventory,
		//CreatedAt will not change
		UpdatedAt: updatedAt,
		//Archived will not change
	}
	return s.repo.UpdateBook(updateBook)
}

func (s *Service) GetBook(id uuid.UUID) (Book, error) {
	return s.repo.GetBookByID(id)
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

func (s *Service) ListBooks(params ListBooksRequest) (PagedBooks, error) {
	itemsTotal, err := s.repo.ListBooksTotals(params.Name, params.MinPrice, params.MaxPrice, params.Archived)
	if err != nil {
		errRepo := ErrResponse{
			Code:    ErrResponseFromRespository.Code,
			Message: ErrResponseFromRespository.Message + err.Error(),
		}
		return PagedBooks{}, errRepo
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
	returnedBooks, err := s.repo.ListBooks(params.Name, params.MinPrice, params.MaxPrice, params.SortBy, params.SortDirection, params.Archived, params.Page, params.PageSize)
	if err != nil {
		errRepo := ErrResponse{
			Code:    ErrResponseFromRespository.Code,
			Message: ErrResponseFromRespository.Message + err.Error(),
		}
		return PagedBooks{}, errRepo
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
