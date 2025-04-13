// Code generated by MockGen. DO NOT EDIT.
// Source: internal/pvz/pg_repository.go

// Package mock_pvz is a generated GoMock package.
package mock_pvz

import (
	context "context"
	reflect "reflect"
	time "time"

	models "github.com/cyansnbrst/pvz-service/internal/models"
	gomock "github.com/golang/mock/gomock"
	uuid "github.com/google/uuid"
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

// AddProduct mocks base method.
func (m *MockRepository) AddProduct(ctx context.Context, productID, pvzID uuid.UUID, productType string) (*models.Product, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddProduct", ctx, productID, pvzID, productType)
	ret0, _ := ret[0].(*models.Product)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddProduct indicates an expected call of AddProduct.
func (mr *MockRepositoryMockRecorder) AddProduct(ctx, productID, pvzID, productType interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddProduct", reflect.TypeOf((*MockRepository)(nil).AddProduct), ctx, productID, pvzID, productType)
}

// CloseLastReception mocks base method.
func (m *MockRepository) CloseLastReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CloseLastReception", ctx, pvzID)
	ret0, _ := ret[0].(*models.Reception)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CloseLastReception indicates an expected call of CloseLastReception.
func (mr *MockRepositoryMockRecorder) CloseLastReception(ctx, pvzID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CloseLastReception", reflect.TypeOf((*MockRepository)(nil).CloseLastReception), ctx, pvzID)
}

// CreatePVZ mocks base method.
func (m *MockRepository) CreatePVZ(ctx context.Context, pvz models.PVZ) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreatePVZ", ctx, pvz)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreatePVZ indicates an expected call of CreatePVZ.
func (mr *MockRepositoryMockRecorder) CreatePVZ(ctx, pvz interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatePVZ", reflect.TypeOf((*MockRepository)(nil).CreatePVZ), ctx, pvz)
}

// CreateReception mocks base method.
func (m *MockRepository) CreateReception(ctx context.Context, receptionID, pvzID uuid.UUID) (*models.Reception, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateReception", ctx, receptionID, pvzID)
	ret0, _ := ret[0].(*models.Reception)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateReception indicates an expected call of CreateReception.
func (mr *MockRepositoryMockRecorder) CreateReception(ctx, receptionID, pvzID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateReception", reflect.TypeOf((*MockRepository)(nil).CreateReception), ctx, receptionID, pvzID)
}

// CreateUser mocks base method.
func (m *MockRepository) CreateUser(ctx context.Context, user models.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUser", ctx, user)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateUser indicates an expected call of CreateUser.
func (mr *MockRepositoryMockRecorder) CreateUser(ctx, user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUser", reflect.TypeOf((*MockRepository)(nil).CreateUser), ctx, user)
}

// DeleteLastProduct mocks base method.
func (m *MockRepository) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteLastProduct", ctx, pvzID)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteLastProduct indicates an expected call of DeleteLastProduct.
func (mr *MockRepositoryMockRecorder) DeleteLastProduct(ctx, pvzID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteLastProduct", reflect.TypeOf((*MockRepository)(nil).DeleteLastProduct), ctx, pvzID)
}

// GetPVZList mocks base method.
func (m *MockRepository) GetPVZList(ctx context.Context) ([]models.PVZ, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPVZList", ctx)
	ret0, _ := ret[0].([]models.PVZ)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPVZList indicates an expected call of GetPVZList.
func (mr *MockRepositoryMockRecorder) GetPVZList(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPVZList", reflect.TypeOf((*MockRepository)(nil).GetPVZList), ctx)
}

// GetPVZs mocks base method.
func (m *MockRepository) GetPVZs(ctx context.Context, startDate, endDate *time.Time, limit, offset uint64) ([]*models.PVZWithReceptions, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPVZs", ctx, startDate, endDate, limit, offset)
	ret0, _ := ret[0].([]*models.PVZWithReceptions)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPVZs indicates an expected call of GetPVZs.
func (mr *MockRepositoryMockRecorder) GetPVZs(ctx, startDate, endDate, limit, offset interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPVZs", reflect.TypeOf((*MockRepository)(nil).GetPVZs), ctx, startDate, endDate, limit, offset)
}

// GetUserByEmail mocks base method.
func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByEmail", ctx, email)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserByEmail indicates an expected call of GetUserByEmail.
func (mr *MockRepositoryMockRecorder) GetUserByEmail(ctx, email interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByEmail", reflect.TypeOf((*MockRepository)(nil).GetUserByEmail), ctx, email)
}
