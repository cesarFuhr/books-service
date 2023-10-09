package book

import "github.com/google/uuid"

type ServiceAPI interface {
	GetBook(id uuid.UUID) (Book, error)
}

type Repository interface {
	GetBookByID(id uuid.UUID) (Book, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetBook(id uuid.UUID) (Book, error) {
	return s.repo.GetBookByID(id)
}
