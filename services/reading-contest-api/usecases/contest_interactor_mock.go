// Code generated by MockGen. DO NOT EDIT.
// Source: contest_interactor.go

// Package usecases is a generated GoMock package.
package usecases

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	domain "github.com/tadoku/tadoku/services/reading-contest-api/domain"
)

// MockContestInteractor is a mock of ContestInteractor interface.
type MockContestInteractor struct {
	ctrl     *gomock.Controller
	recorder *MockContestInteractorMockRecorder
}

// MockContestInteractorMockRecorder is the mock recorder for MockContestInteractor.
type MockContestInteractorMockRecorder struct {
	mock *MockContestInteractor
}

// NewMockContestInteractor creates a new mock instance.
func NewMockContestInteractor(ctrl *gomock.Controller) *MockContestInteractor {
	mock := &MockContestInteractor{ctrl: ctrl}
	mock.recorder = &MockContestInteractorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockContestInteractor) EXPECT() *MockContestInteractorMockRecorder {
	return m.recorder
}

// CreateContest mocks base method.
func (m *MockContestInteractor) CreateContest(contest domain.Contest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateContest", contest)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateContest indicates an expected call of CreateContest.
func (mr *MockContestInteractorMockRecorder) CreateContest(contest interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateContest", reflect.TypeOf((*MockContestInteractor)(nil).CreateContest), contest)
}

// Find mocks base method.
func (m *MockContestInteractor) Find(contestID uint64) (*domain.Contest, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Find", contestID)
	ret0, _ := ret[0].(*domain.Contest)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Find indicates an expected call of Find.
func (mr *MockContestInteractorMockRecorder) Find(contestID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Find", reflect.TypeOf((*MockContestInteractor)(nil).Find), contestID)
}

// Recent mocks base method.
func (m *MockContestInteractor) Recent(count int) ([]domain.Contest, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Recent", count)
	ret0, _ := ret[0].([]domain.Contest)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Recent indicates an expected call of Recent.
func (mr *MockContestInteractorMockRecorder) Recent(count interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Recent", reflect.TypeOf((*MockContestInteractor)(nil).Recent), count)
}

// Stats mocks base method.
func (m *MockContestInteractor) Stats(contestID uint64) (*domain.ContestStats, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stats", contestID)
	ret0, _ := ret[0].(*domain.ContestStats)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Stats indicates an expected call of Stats.
func (mr *MockContestInteractorMockRecorder) Stats(contestID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stats", reflect.TypeOf((*MockContestInteractor)(nil).Stats), contestID)
}

// UpdateContest mocks base method.
func (m *MockContestInteractor) UpdateContest(contest domain.Contest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateContest", contest)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateContest indicates an expected call of UpdateContest.
func (mr *MockContestInteractorMockRecorder) UpdateContest(contest interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateContest", reflect.TypeOf((*MockContestInteractor)(nil).UpdateContest), contest)
}
