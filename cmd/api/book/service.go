package book

import "github.com/google/uuid"

type ServiceAPI interface {
	ArchiveStatusBook(id uuid.UUID) (Book, error)
	GetBook(id uuid.UUID) (Book, error)
}

type Repository interface {
	ArchiveStatusBook(id uuid.UUID, archived bool) (Book, error)
	GetBookByID(id uuid.UUID) (Book, error)
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
