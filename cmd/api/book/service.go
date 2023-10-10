package book

import (
	"time"

	"github.com/google/uuid"
)

type ServiceAPI interface {
	ArchiveStatusBook(id uuid.UUID) (Book, error)
	GetBook(id uuid.UUID) (Book, error)
	UpdateBook(bookEntry Book, id uuid.UUID) (Book, error)
}

type Repository interface {
	ArchiveStatusBook(id uuid.UUID, archived bool) (Book, error)
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

func (s *Service) GetBook(id uuid.UUID) (Book, error) {
	return s.repo.GetBookByID(id)
}

func (s *Service) UpdateBook(bookEntry Book, id uuid.UUID) (Book, error) {
	bookEntry.ID = id
	bookEntry.UpdatedAt = time.Now().UTC().Round(time.Millisecond) //Atribute a new updating time to the new entry.
	return s.repo.UpdateBook(bookEntry)
}
