// Code generated by MockGen. DO NOT EDIT.
// Source: storage.go

// Package mocks_storage is a generated GoMock package.
package mocks_storage

import (
	context "context"
	reflect "reflect"

	models "github.com/Grino777/sso/internal/domain/models"
	gomock "github.com/golang/mock/gomock"
)

// MockStorage is a mock of Storage interface.
type MockStorage struct {
	ctrl     *gomock.Controller
	recorder *MockStorageMockRecorder
}

// MockStorageMockRecorder is the mock recorder for MockStorage.
type MockStorageMockRecorder struct {
	mock *MockStorage
}

// NewMockStorage creates a new mock instance.
func NewMockStorage(ctrl *gomock.Controller) *MockStorage {
	mock := &MockStorage{ctrl: ctrl}
	mock.recorder = &MockStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorage) EXPECT() *MockStorageMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockStorage) Close(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockStorageMockRecorder) Close(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockStorage)(nil).Close), ctx)
}

// Connect mocks base method.
func (m *MockStorage) Connect(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Connect", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Connect indicates an expected call of Connect.
func (mr *MockStorageMockRecorder) Connect(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Connect", reflect.TypeOf((*MockStorage)(nil).Connect), ctx)
}

// DeleteRefreshToken mocks base method.
func (m *MockStorage) DeleteRefreshToken(ctx context.Context, userID uint64, appID uint32, token models.Token) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteRefreshToken", ctx, userID, appID, token)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteRefreshToken indicates an expected call of DeleteRefreshToken.
func (mr *MockStorageMockRecorder) DeleteRefreshToken(ctx, userID, appID, token interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteRefreshToken", reflect.TypeOf((*MockStorage)(nil).DeleteRefreshToken), ctx, userID, appID, token)
}

// GetApp mocks base method.
func (m *MockStorage) GetApp(ctx context.Context, appID uint32) (models.App, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetApp", ctx, appID)
	ret0, _ := ret[0].(models.App)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetApp indicates an expected call of GetApp.
func (mr *MockStorageMockRecorder) GetApp(ctx, appID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApp", reflect.TypeOf((*MockStorage)(nil).GetApp), ctx, appID)
}

// GetUser mocks base method.
func (m *MockStorage) GetUser(ctx context.Context, username string) (models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUser", ctx, username)
	ret0, _ := ret[0].(models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUser indicates an expected call of GetUser.
func (mr *MockStorageMockRecorder) GetUser(ctx, username interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUser", reflect.TypeOf((*MockStorage)(nil).GetUser), ctx, username)
}

// IsAdmin mocks base method.
func (m *MockStorage) IsAdmin(ctx context.Context, username string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsAdmin", ctx, username)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsAdmin indicates an expected call of IsAdmin.
func (mr *MockStorageMockRecorder) IsAdmin(ctx, username interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsAdmin", reflect.TypeOf((*MockStorage)(nil).IsAdmin), ctx, username)
}

// SaveRefreshToken mocks base method.
func (m *MockStorage) SaveRefreshToken(ctx context.Context, userID uint64, appID uint32, token models.Token) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveRefreshToken", ctx, userID, appID, token)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveRefreshToken indicates an expected call of SaveRefreshToken.
func (mr *MockStorageMockRecorder) SaveRefreshToken(ctx, userID, appID, token interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveRefreshToken", reflect.TypeOf((*MockStorage)(nil).SaveRefreshToken), ctx, userID, appID, token)
}

// SaveUser mocks base method.
func (m *MockStorage) SaveUser(ctx context.Context, user, passHash string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveUser", ctx, user, passHash)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveUser indicates an expected call of SaveUser.
func (mr *MockStorageMockRecorder) SaveUser(ctx, user, passHash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveUser", reflect.TypeOf((*MockStorage)(nil).SaveUser), ctx, user, passHash)
}

// MockStorageUserProvider is a mock of StorageUserProvider interface.
type MockStorageUserProvider struct {
	ctrl     *gomock.Controller
	recorder *MockStorageUserProviderMockRecorder
}

// MockStorageUserProviderMockRecorder is the mock recorder for MockStorageUserProvider.
type MockStorageUserProviderMockRecorder struct {
	mock *MockStorageUserProvider
}

// NewMockStorageUserProvider creates a new mock instance.
func NewMockStorageUserProvider(ctrl *gomock.Controller) *MockStorageUserProvider {
	mock := &MockStorageUserProvider{ctrl: ctrl}
	mock.recorder = &MockStorageUserProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorageUserProvider) EXPECT() *MockStorageUserProviderMockRecorder {
	return m.recorder
}

// GetUser mocks base method.
func (m *MockStorageUserProvider) GetUser(ctx context.Context, username string) (models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUser", ctx, username)
	ret0, _ := ret[0].(models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUser indicates an expected call of GetUser.
func (mr *MockStorageUserProviderMockRecorder) GetUser(ctx, username interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUser", reflect.TypeOf((*MockStorageUserProvider)(nil).GetUser), ctx, username)
}

// IsAdmin mocks base method.
func (m *MockStorageUserProvider) IsAdmin(ctx context.Context, username string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsAdmin", ctx, username)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsAdmin indicates an expected call of IsAdmin.
func (mr *MockStorageUserProviderMockRecorder) IsAdmin(ctx, username interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsAdmin", reflect.TypeOf((*MockStorageUserProvider)(nil).IsAdmin), ctx, username)
}

// SaveUser mocks base method.
func (m *MockStorageUserProvider) SaveUser(ctx context.Context, user, passHash string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveUser", ctx, user, passHash)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveUser indicates an expected call of SaveUser.
func (mr *MockStorageUserProviderMockRecorder) SaveUser(ctx, user, passHash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveUser", reflect.TypeOf((*MockStorageUserProvider)(nil).SaveUser), ctx, user, passHash)
}

// MockStorageAppProvider is a mock of StorageAppProvider interface.
type MockStorageAppProvider struct {
	ctrl     *gomock.Controller
	recorder *MockStorageAppProviderMockRecorder
}

// MockStorageAppProviderMockRecorder is the mock recorder for MockStorageAppProvider.
type MockStorageAppProviderMockRecorder struct {
	mock *MockStorageAppProvider
}

// NewMockStorageAppProvider creates a new mock instance.
func NewMockStorageAppProvider(ctrl *gomock.Controller) *MockStorageAppProvider {
	mock := &MockStorageAppProvider{ctrl: ctrl}
	mock.recorder = &MockStorageAppProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorageAppProvider) EXPECT() *MockStorageAppProviderMockRecorder {
	return m.recorder
}

// GetApp mocks base method.
func (m *MockStorageAppProvider) GetApp(ctx context.Context, appID uint32) (models.App, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetApp", ctx, appID)
	ret0, _ := ret[0].(models.App)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetApp indicates an expected call of GetApp.
func (mr *MockStorageAppProviderMockRecorder) GetApp(ctx, appID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApp", reflect.TypeOf((*MockStorageAppProvider)(nil).GetApp), ctx, appID)
}

// MockStorageTokenProvider is a mock of StorageTokenProvider interface.
type MockStorageTokenProvider struct {
	ctrl     *gomock.Controller
	recorder *MockStorageTokenProviderMockRecorder
}

// MockStorageTokenProviderMockRecorder is the mock recorder for MockStorageTokenProvider.
type MockStorageTokenProviderMockRecorder struct {
	mock *MockStorageTokenProvider
}

// NewMockStorageTokenProvider creates a new mock instance.
func NewMockStorageTokenProvider(ctrl *gomock.Controller) *MockStorageTokenProvider {
	mock := &MockStorageTokenProvider{ctrl: ctrl}
	mock.recorder = &MockStorageTokenProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorageTokenProvider) EXPECT() *MockStorageTokenProviderMockRecorder {
	return m.recorder
}

// DeleteRefreshToken mocks base method.
func (m *MockStorageTokenProvider) DeleteRefreshToken(ctx context.Context, userID uint64, appID uint32, token models.Token) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteRefreshToken", ctx, userID, appID, token)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteRefreshToken indicates an expected call of DeleteRefreshToken.
func (mr *MockStorageTokenProviderMockRecorder) DeleteRefreshToken(ctx, userID, appID, token interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteRefreshToken", reflect.TypeOf((*MockStorageTokenProvider)(nil).DeleteRefreshToken), ctx, userID, appID, token)
}

// SaveRefreshToken mocks base method.
func (m *MockStorageTokenProvider) SaveRefreshToken(ctx context.Context, userID uint64, appID uint32, token models.Token) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveRefreshToken", ctx, userID, appID, token)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveRefreshToken indicates an expected call of SaveRefreshToken.
func (mr *MockStorageTokenProviderMockRecorder) SaveRefreshToken(ctx, userID, appID, token interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveRefreshToken", reflect.TypeOf((*MockStorageTokenProvider)(nil).SaveRefreshToken), ctx, userID, appID, token)
}

// MockConnector is a mock of Connector interface.
type MockConnector struct {
	ctrl     *gomock.Controller
	recorder *MockConnectorMockRecorder
}

// MockConnectorMockRecorder is the mock recorder for MockConnector.
type MockConnectorMockRecorder struct {
	mock *MockConnector
}

// NewMockConnector creates a new mock instance.
func NewMockConnector(ctrl *gomock.Controller) *MockConnector {
	mock := &MockConnector{ctrl: ctrl}
	mock.recorder = &MockConnectorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConnector) EXPECT() *MockConnectorMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockConnector) Close(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockConnectorMockRecorder) Close(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockConnector)(nil).Close), ctx)
}

// Connect mocks base method.
func (m *MockConnector) Connect(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Connect", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Connect indicates an expected call of Connect.
func (mr *MockConnectorMockRecorder) Connect(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Connect", reflect.TypeOf((*MockConnector)(nil).Connect), ctx)
}
