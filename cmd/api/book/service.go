package book

import (
	"math"
	"time"

	"github.com/google/uuid"
)

type ServiceAPI interface {
	ArchiveBook(id uuid.UUID) (Book, error)
	CreateBook(bookEntry Book) (Book, error)
	GetBook(id uuid.UUID) (Book, error)
	ListBooks(params ListBooksRequest) (PagedBooks, error)
	UpdateBook(bookEntry Book, id uuid.UUID) (Book, error)
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

func (s *Service) CreateBook(bookEntry Book) (Book, error) {
	bookEntry.ID = uuid.New()                                      //Atribute an ID to the entry
	bookEntry.CreatedAt = time.Now().UTC().Round(time.Millisecond) //Atribute creating and updating time to the new entry. UpdateAt can change later.
	bookEntry.UpdatedAt = bookEntry.CreatedAt
	return s.repo.CreateBook(bookEntry)
}

func (s *Service) GetBook(id uuid.UUID) (Book, error) {
	return s.repo.GetBookByID(id)
}

type PagedBooks struct {
	PageCurrent int    `json:"page_current"`
	PageTotal   int    `json:"page_total"`
	PageSize    int    `json:"page_size"`
	ItemsTotal  int    `json:"items_total"`
	Results     []Book `json:"results"`
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

func (s *Service) UpdateBook(bookEntry Book, id uuid.UUID) (Book, error) {
	bookEntry.ID = id
	bookEntry.UpdatedAt = time.Now().UTC().Round(time.Millisecond) //Atribute a new updating time to the new entry.
	return s.repo.UpdateBook(bookEntry)
}

/*Calculates the pagination.*/
func pagination(page, pageSize, itemsTotal int) (pagesTotal int, err error) {
	pagesTotal = int(math.Ceil(float64(itemsTotal) / float64(pageSize)))
	if page > pagesTotal {
		return 0, ErrResponseQueryPageOutOfRange
	}

	return pagesTotal, nil
}
