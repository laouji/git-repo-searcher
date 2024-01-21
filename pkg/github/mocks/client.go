// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/github/client.go
//
// Generated by this command:
//
//	mockgen -destination=pkg/github/mocks/client.go -source=pkg/github/client.go
//

// Package mock_github is a generated GoMock package.
package mock_github

import (
	context "context"
	reflect "reflect"

	github "github.com/laouji/git-repo-searcher/pkg/github"
	gomock "go.uber.org/mock/gomock"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// FetchAttribute mocks base method.
func (m *MockClient) FetchAttribute(ctx context.Context, url string) (map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchAttribute", ctx, url)
	ret0, _ := ret[0].(map[string]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchAttribute indicates an expected call of FetchAttribute.
func (mr *MockClientMockRecorder) FetchAttribute(ctx, url any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchAttribute", reflect.TypeOf((*MockClient)(nil).FetchAttribute), ctx, url)
}

// ListPublicEvents mocks base method.
func (m *MockClient) ListPublicEvents(ctx context.Context, limit, offset int) ([]github.Event, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListPublicEvents", ctx, limit, offset)
	ret0, _ := ret[0].([]github.Event)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListPublicEvents indicates an expected call of ListPublicEvents.
func (mr *MockClientMockRecorder) ListPublicEvents(ctx, limit, offset any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListPublicEvents", reflect.TypeOf((*MockClient)(nil).ListPublicEvents), ctx, limit, offset)
}

// ListPublicRepos mocks base method.
func (m *MockClient) ListPublicRepos(ctx context.Context, since int64) ([]github.Repository, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListPublicRepos", ctx, since)
	ret0, _ := ret[0].([]github.Repository)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListPublicRepos indicates an expected call of ListPublicRepos.
func (mr *MockClientMockRecorder) ListPublicRepos(ctx, since any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListPublicRepos", reflect.TypeOf((*MockClient)(nil).ListPublicRepos), ctx, since)
}

// SetAccessToken mocks base method.
func (m *MockClient) SetAccessToken(ctx context.Context) (github.Token, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetAccessToken", ctx)
	ret0, _ := ret[0].(github.Token)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SetAccessToken indicates an expected call of SetAccessToken.
func (mr *MockClientMockRecorder) SetAccessToken(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetAccessToken", reflect.TypeOf((*MockClient)(nil).SetAccessToken), ctx)
}
