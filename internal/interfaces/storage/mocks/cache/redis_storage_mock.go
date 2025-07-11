// Code generated by MockGen. DO NOT EDIT.
// Source: cache.go

// Package mock_cache is a generated GoMock package.
package mock_cache

import (
	context "context"
	reflect "reflect"

	models "github.com/Grino777/sso/internal/domain/models"
	gomock "github.com/golang/mock/gomock"
)

// MockCacheStorage is a mock of CacheStorage interface.
type MockCacheStorage struct {
	ctrl     *gomock.Controller
	recorder *MockCacheStorageMockRecorder
}

// MockCacheStorageMockRecorder is the mock recorder for MockCacheStorage.
type MockCacheStorageMockRecorder struct {
	mock *MockCacheStorage
}

// NewMockCacheStorage creates a new mock instance.
func NewMockCacheStorage(ctrl *gomock.Controller) *MockCacheStorage {
	mock := &MockCacheStorage{ctrl: ctrl}
	mock.recorder = &MockCacheStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCacheStorage) EXPECT() *MockCacheStorageMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockCacheStorage) Close(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockCacheStorageMockRecorder) Close(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockCacheStorage)(nil).Close), ctx)
}

// Connect mocks base method.
func (m *MockCacheStorage) Connect(ctx context.Context, errChan chan<- error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Connect", ctx, errChan)
	ret0, _ := ret[0].(error)
	return ret0
}

// Connect indicates an expected call of Connect.
func (mr *MockCacheStorageMockRecorder) Connect(ctx, errChan interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Connect", reflect.TypeOf((*MockCacheStorage)(nil).Connect), ctx, errChan)
}

// GetApp mocks base method.
func (m *MockCacheStorage) GetApp(ctx context.Context, appID uint32) (models.App, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetApp", ctx, appID)
	ret0, _ := ret[0].(models.App)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetApp indicates an expected call of GetApp.
func (mr *MockCacheStorageMockRecorder) GetApp(ctx, appID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApp", reflect.TypeOf((*MockCacheStorage)(nil).GetApp), ctx, appID)
}

// GetUser mocks base method.
func (m *MockCacheStorage) GetUser(ctx context.Context, username string, appID uint32) (models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUser", ctx, username, appID)
	ret0, _ := ret[0].(models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUser indicates an expected call of GetUser.
func (mr *MockCacheStorageMockRecorder) GetUser(ctx, username, appID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUser", reflect.TypeOf((*MockCacheStorage)(nil).GetUser), ctx, username, appID)
}

// IsAdmin mocks base method.
func (m *MockCacheStorage) IsAdmin(ctx context.Context, user models.User, app models.App) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsAdmin", ctx, user, app)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsAdmin indicates an expected call of IsAdmin.
func (mr *MockCacheStorageMockRecorder) IsAdmin(ctx, user, app interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsAdmin", reflect.TypeOf((*MockCacheStorage)(nil).IsAdmin), ctx, user, app)
}

// SaveApp mocks base method.
func (m *MockCacheStorage) SaveApp(ctx context.Context, app models.App) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveApp", ctx, app)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveApp indicates an expected call of SaveApp.
func (mr *MockCacheStorageMockRecorder) SaveApp(ctx, app interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveApp", reflect.TypeOf((*MockCacheStorage)(nil).SaveApp), ctx, app)
}

// SaveUser mocks base method.
func (m *MockCacheStorage) SaveUser(ctx context.Context, user models.User, appID uint32) (models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveUser", ctx, user, appID)
	ret0, _ := ret[0].(models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SaveUser indicates an expected call of SaveUser.
func (mr *MockCacheStorageMockRecorder) SaveUser(ctx, user, appID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveUser", reflect.TypeOf((*MockCacheStorage)(nil).SaveUser), ctx, user, appID)
}

// MockCacheUserProvider is a mock of CacheUserProvider interface.
type MockCacheUserProvider struct {
	ctrl     *gomock.Controller
	recorder *MockCacheUserProviderMockRecorder
}

// MockCacheUserProviderMockRecorder is the mock recorder for MockCacheUserProvider.
type MockCacheUserProviderMockRecorder struct {
	mock *MockCacheUserProvider
}

// NewMockCacheUserProvider creates a new mock instance.
func NewMockCacheUserProvider(ctrl *gomock.Controller) *MockCacheUserProvider {
	mock := &MockCacheUserProvider{ctrl: ctrl}
	mock.recorder = &MockCacheUserProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCacheUserProvider) EXPECT() *MockCacheUserProviderMockRecorder {
	return m.recorder
}

// GetUser mocks base method.
func (m *MockCacheUserProvider) GetUser(ctx context.Context, username string, appID uint32) (models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUser", ctx, username, appID)
	ret0, _ := ret[0].(models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUser indicates an expected call of GetUser.
func (mr *MockCacheUserProviderMockRecorder) GetUser(ctx, username, appID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUser", reflect.TypeOf((*MockCacheUserProvider)(nil).GetUser), ctx, username, appID)
}

// IsAdmin mocks base method.
func (m *MockCacheUserProvider) IsAdmin(ctx context.Context, user models.User, app models.App) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsAdmin", ctx, user, app)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsAdmin indicates an expected call of IsAdmin.
func (mr *MockCacheUserProviderMockRecorder) IsAdmin(ctx, user, app interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsAdmin", reflect.TypeOf((*MockCacheUserProvider)(nil).IsAdmin), ctx, user, app)
}

// SaveUser mocks base method.
func (m *MockCacheUserProvider) SaveUser(ctx context.Context, user models.User, appID uint32) (models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveUser", ctx, user, appID)
	ret0, _ := ret[0].(models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SaveUser indicates an expected call of SaveUser.
func (mr *MockCacheUserProviderMockRecorder) SaveUser(ctx, user, appID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveUser", reflect.TypeOf((*MockCacheUserProvider)(nil).SaveUser), ctx, user, appID)
}

// MockCacheAppProvider is a mock of CacheAppProvider interface.
type MockCacheAppProvider struct {
	ctrl     *gomock.Controller
	recorder *MockCacheAppProviderMockRecorder
}

// MockCacheAppProviderMockRecorder is the mock recorder for MockCacheAppProvider.
type MockCacheAppProviderMockRecorder struct {
	mock *MockCacheAppProvider
}

// NewMockCacheAppProvider creates a new mock instance.
func NewMockCacheAppProvider(ctrl *gomock.Controller) *MockCacheAppProvider {
	mock := &MockCacheAppProvider{ctrl: ctrl}
	mock.recorder = &MockCacheAppProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCacheAppProvider) EXPECT() *MockCacheAppProviderMockRecorder {
	return m.recorder
}

// GetApp mocks base method.
func (m *MockCacheAppProvider) GetApp(ctx context.Context, appID uint32) (models.App, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetApp", ctx, appID)
	ret0, _ := ret[0].(models.App)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetApp indicates an expected call of GetApp.
func (mr *MockCacheAppProviderMockRecorder) GetApp(ctx, appID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApp", reflect.TypeOf((*MockCacheAppProvider)(nil).GetApp), ctx, appID)
}

// SaveApp mocks base method.
func (m *MockCacheAppProvider) SaveApp(ctx context.Context, app models.App) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveApp", ctx, app)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveApp indicates an expected call of SaveApp.
func (mr *MockCacheAppProviderMockRecorder) SaveApp(ctx, app interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveApp", reflect.TypeOf((*MockCacheAppProvider)(nil).SaveApp), ctx, app)
}

// MockCacheConnector is a mock of CacheConnector interface.
type MockCacheConnector struct {
	ctrl     *gomock.Controller
	recorder *MockCacheConnectorMockRecorder
}

// MockCacheConnectorMockRecorder is the mock recorder for MockCacheConnector.
type MockCacheConnectorMockRecorder struct {
	mock *MockCacheConnector
}

// NewMockCacheConnector creates a new mock instance.
func NewMockCacheConnector(ctrl *gomock.Controller) *MockCacheConnector {
	mock := &MockCacheConnector{ctrl: ctrl}
	mock.recorder = &MockCacheConnectorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCacheConnector) EXPECT() *MockCacheConnectorMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockCacheConnector) Close(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockCacheConnectorMockRecorder) Close(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockCacheConnector)(nil).Close), ctx)
}

// Connect mocks base method.
func (m *MockCacheConnector) Connect(ctx context.Context, errChan chan<- error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Connect", ctx, errChan)
	ret0, _ := ret[0].(error)
	return ret0
}

// Connect indicates an expected call of Connect.
func (mr *MockCacheConnectorMockRecorder) Connect(ctx, errChan interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Connect", reflect.TypeOf((*MockCacheConnector)(nil).Connect), ctx, errChan)
}
