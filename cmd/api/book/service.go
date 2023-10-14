package book

import (
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type ServiceAPI interface {
	ArchiveBook(id uuid.UUID) (Book, error)
	CreateBook(bookEntry Book) (Book, error)
	GetBook(id uuid.UUID) (Book, error)
	ListBooks(query url.Values) (PagedBooks, error)
	UpdateBook(bookEntry Book, id uuid.UUID) (Book, error)
}

type Repository interface {
	SetBookArchiveStatus(id uuid.UUID, archived bool) (Book, error)
	CountRows(name string, minPrice32, maxPrice32 float32, archived bool) (int, error)
	CreateBook(bookEntry Book) (Book, error)
	GetBookByID(id uuid.UUID) (Book, error)
	ListBooks(name string, minPrice32, maxPrice32 float32, sortBy, sortDirection string, archived bool, page, pageSize int) ([]Book, error)
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

func (s *Service) ListBooks(query url.Values) (PagedBooks, error) {
	name := query.Get("name")

	var minPrice32 float32
	minPriceStr := query.Get("min_price")
	if minPriceStr != "" {
		minPrice64, err := strconv.ParseFloat(minPriceStr, 32)
		if err != nil {
			return PagedBooks{}, ErrResponseQueryPriceInvalidFormat
		}
		minPrice32 = float32(minPrice64)
	} else {
		minPrice32 = 0
	}

	var maxPrice32 float32
	maxPriceStr := query.Get("max_price")
	if maxPriceStr != "" {
		maxPrice64, err := strconv.ParseFloat(maxPriceStr, 32)
		if err != nil {
			return PagedBooks{}, ErrResponseQueryPriceInvalidFormat
		}
		maxPrice32 = float32(maxPrice64)
	} else {
		maxPrice32 = 9999.99 //max value to field price on db, set to: numeric(6,2)
	}

	sortBy, sortDirection, valid := extractOrderParams(query)
	if !valid {
		return PagedBooks{}, ErrResponseQuerySortByInvalid
	}

	archived := false
	archivedStr := query.Get("archived")
	if archivedStr == "true" {
		archived = true
	}

	itemsTotal, err := s.repo.CountRows(name, minPrice32, maxPrice32, archived)
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

	pagesTotal, page, pageSize, err := pagination(query, itemsTotal)
	if err != nil {
		return PagedBooks{}, err
	}

	//Ask filtered list to db:
	returnedBooks, err := s.repo.ListBooks(name, minPrice32, maxPrice32, sortBy, sortDirection, archived, page, pageSize)
	if err != nil {
		errRepo := ErrResponse{
			Code:    ErrResponseFromRespository.Code,
			Message: ErrResponseFromRespository.Message + err.Error(),
		}
		return PagedBooks{}, errRepo
	}

	pageOfBooksList := PagedBooks{
		PageCurrent: page,
		PageTotal:   pagesTotal,
		PageSize:    pageSize,
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
