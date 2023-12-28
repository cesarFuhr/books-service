// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/books-service/cmd/api/book (interfaces: Repository)
//
// Generated by this command:
//
//	mockgen.exe -destination ./mocks/mocks.go -package book . Repository
//
// Package book is a generated GoMock package.
package book

import (
	context "context"
	reflect "reflect"

	book "github.com/books-service/cmd/api/book"
	uuid "github.com/google/uuid"
	gomock "go.uber.org/mock/gomock"
)

// MockRepository is a mock of Repository interface.
type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository.
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository creates a new mock instance.
func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

// CreateBook mocks base method.
func (m *MockRepository) CreateBook(arg0 context.Context, arg1 book.Book) (book.Book, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateBook", arg0, arg1)
	ret0, _ := ret[0].(book.Book)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateBook indicates an expected call of CreateBook.
func (mr *MockRepositoryMockRecorder) CreateBook(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateBook", reflect.TypeOf((*MockRepository)(nil).CreateBook), arg0, arg1)
}

// GetBookByID mocks base method.
func (m *MockRepository) GetBookByID(arg0 uuid.UUID) (book.Book, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBookByID", arg0)
	ret0, _ := ret[0].(book.Book)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBookByID indicates an expected call of GetBookByID.
func (mr *MockRepositoryMockRecorder) GetBookByID(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBookByID", reflect.TypeOf((*MockRepository)(nil).GetBookByID), arg0)
}

// ListBooks mocks base method.
func (m *MockRepository) ListBooks(arg0 string, arg1, arg2 float32, arg3, arg4 string, arg5 bool, arg6, arg7 int) ([]book.Book, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListBooks", arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
	ret0, _ := ret[0].([]book.Book)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListBooks indicates an expected call of ListBooks.
func (mr *MockRepositoryMockRecorder) ListBooks(arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListBooks", reflect.TypeOf((*MockRepository)(nil).ListBooks), arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
}

// ListBooksTotals mocks base method.
func (m *MockRepository) ListBooksTotals(arg0 string, arg1, arg2 float32, arg3 bool) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListBooksTotals", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListBooksTotals indicates an expected call of ListBooksTotals.
func (mr *MockRepositoryMockRecorder) ListBooksTotals(arg0, arg1, arg2, arg3 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListBooksTotals", reflect.TypeOf((*MockRepository)(nil).ListBooksTotals), arg0, arg1, arg2, arg3)
}

// SetBookArchiveStatus mocks base method.
func (m *MockRepository) SetBookArchiveStatus(arg0 uuid.UUID, arg1 bool) (book.Book, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetBookArchiveStatus", arg0, arg1)
	ret0, _ := ret[0].(book.Book)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SetBookArchiveStatus indicates an expected call of SetBookArchiveStatus.
func (mr *MockRepositoryMockRecorder) SetBookArchiveStatus(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetBookArchiveStatus", reflect.TypeOf((*MockRepository)(nil).SetBookArchiveStatus), arg0, arg1)
}

// UpdateBook mocks base method.
func (m *MockRepository) UpdateBook(arg0 book.Book) (book.Book, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateBook", arg0)
	ret0, _ := ret[0].(book.Book)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateBook indicates an expected call of UpdateBook.
func (mr *MockRepositoryMockRecorder) UpdateBook(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateBook", reflect.TypeOf((*MockRepository)(nil).UpdateBook), arg0)
}
