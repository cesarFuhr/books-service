package book

import (
	"time"

	"github.com/google/uuid"
)

type ServiceAPI interface {
	ArchiveStatusBook(id uuid.UUID) (Book, error)
	CreateBook(bookEntry Book) (Book, error)
	GetBook(id uuid.UUID) (Book, error)
	UpdateBook(bookEntry Book, id uuid.UUID) (Book, error)
}

type Repository interface {
	ArchiveStatusBook(id uuid.UUID, archived bool) (Book, error)
	CreateBook(bookEntry Book) (Book, error)
	GetBookByID(id uuid.UUID) (Book, error)
	UpdateBook(bookEntry Book) (Book, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ArchiveStatusBook(id uuid.UUID) (Book, error) {
	archived := true
	return s.repo.ArchiveStatusBook(id, archived)
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

func (s *Service) UpdateBook(bookEntry Book, id uuid.UUID) (Book, error) {
	bookEntry.ID = id
	bookEntry.UpdatedAt = time.Now().UTC().Round(time.Millisecond) //Atribute a new updating time to the new entry.
	return s.repo.UpdateBook(bookEntry)
}
