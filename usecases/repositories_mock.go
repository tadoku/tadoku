// Code generated by MockGen. DO NOT EDIT.
// Source: repositories.go

// Package usecases is a generated GoMock package.
package usecases

import (
	gomock "github.com/golang/mock/gomock"
	domain "github.com/tadoku/api/domain"
	reflect "reflect"
)

// MockUserRepository is a mock of UserRepository interface
type MockUserRepository struct {
	ctrl     *gomock.Controller
	recorder *MockUserRepositoryMockRecorder
}

// MockUserRepositoryMockRecorder is the mock recorder for MockUserRepository
type MockUserRepositoryMockRecorder struct {
	mock *MockUserRepository
}

// NewMockUserRepository creates a new mock instance
func NewMockUserRepository(ctrl *gomock.Controller) *MockUserRepository {
	mock := &MockUserRepository{ctrl: ctrl}
	mock.recorder = &MockUserRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockUserRepository) EXPECT() *MockUserRepositoryMockRecorder {
	return m.recorder
}

// Store mocks base method
func (m *MockUserRepository) Store(user *domain.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Store", user)
	ret0, _ := ret[0].(error)
	return ret0
}

// Store indicates an expected call of Store
func (mr *MockUserRepositoryMockRecorder) Store(user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Store", reflect.TypeOf((*MockUserRepository)(nil).Store), user)
}

// UpdatePassword mocks base method
func (m *MockUserRepository) UpdatePassword(user *domain.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdatePassword", user)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdatePassword indicates an expected call of UpdatePassword
func (mr *MockUserRepositoryMockRecorder) UpdatePassword(user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatePassword", reflect.TypeOf((*MockUserRepository)(nil).UpdatePassword), user)
}

// FindByID mocks base method
func (m *MockUserRepository) FindByID(id uint64) (domain.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindByID", id)
	ret0, _ := ret[0].(domain.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindByID indicates an expected call of FindByID
func (mr *MockUserRepositoryMockRecorder) FindByID(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByID", reflect.TypeOf((*MockUserRepository)(nil).FindByID), id)
}

// FindByEmail mocks base method
func (m *MockUserRepository) FindByEmail(email string) (domain.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindByEmail", email)
	ret0, _ := ret[0].(domain.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindByEmail indicates an expected call of FindByEmail
func (mr *MockUserRepositoryMockRecorder) FindByEmail(email interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByEmail", reflect.TypeOf((*MockUserRepository)(nil).FindByEmail), email)
}

// MockContestRepository is a mock of ContestRepository interface
type MockContestRepository struct {
	ctrl     *gomock.Controller
	recorder *MockContestRepositoryMockRecorder
}

// MockContestRepositoryMockRecorder is the mock recorder for MockContestRepository
type MockContestRepositoryMockRecorder struct {
	mock *MockContestRepository
}

// NewMockContestRepository creates a new mock instance
func NewMockContestRepository(ctrl *gomock.Controller) *MockContestRepository {
	mock := &MockContestRepository{ctrl: ctrl}
	mock.recorder = &MockContestRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockContestRepository) EXPECT() *MockContestRepositoryMockRecorder {
	return m.recorder
}

// Store mocks base method
func (m *MockContestRepository) Store(contest *domain.Contest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Store", contest)
	ret0, _ := ret[0].(error)
	return ret0
}

// Store indicates an expected call of Store
func (mr *MockContestRepositoryMockRecorder) Store(contest interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Store", reflect.TypeOf((*MockContestRepository)(nil).Store), contest)
}

// GetOpenContests mocks base method
func (m *MockContestRepository) GetOpenContests() ([]uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOpenContests")
	ret0, _ := ret[0].([]uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOpenContests indicates an expected call of GetOpenContests
func (mr *MockContestRepositoryMockRecorder) GetOpenContests() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOpenContests", reflect.TypeOf((*MockContestRepository)(nil).GetOpenContests))
}

// GetRunningContests mocks base method
func (m *MockContestRepository) GetRunningContests() ([]uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRunningContests")
	ret0, _ := ret[0].([]uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRunningContests indicates an expected call of GetRunningContests
func (mr *MockContestRepositoryMockRecorder) GetRunningContests() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRunningContests", reflect.TypeOf((*MockContestRepository)(nil).GetRunningContests))
}

// FindAll mocks base method
func (m *MockContestRepository) FindAll() ([]domain.Contest, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindAll")
	ret0, _ := ret[0].([]domain.Contest)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindAll indicates an expected call of FindAll
func (mr *MockContestRepositoryMockRecorder) FindAll() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindAll", reflect.TypeOf((*MockContestRepository)(nil).FindAll))
}

// FindRecent mocks base method
func (m *MockContestRepository) FindRecent(count int) ([]domain.Contest, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindRecent", count)
	ret0, _ := ret[0].([]domain.Contest)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindRecent indicates an expected call of FindRecent
func (mr *MockContestRepositoryMockRecorder) FindRecent(count interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindRecent", reflect.TypeOf((*MockContestRepository)(nil).FindRecent), count)
}

// FindByID mocks base method
func (m *MockContestRepository) FindByID(id uint64) (domain.Contest, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindByID", id)
	ret0, _ := ret[0].(domain.Contest)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindByID indicates an expected call of FindByID
func (mr *MockContestRepositoryMockRecorder) FindByID(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByID", reflect.TypeOf((*MockContestRepository)(nil).FindByID), id)
}

// MockContestLogRepository is a mock of ContestLogRepository interface
type MockContestLogRepository struct {
	ctrl     *gomock.Controller
	recorder *MockContestLogRepositoryMockRecorder
}

// MockContestLogRepositoryMockRecorder is the mock recorder for MockContestLogRepository
type MockContestLogRepositoryMockRecorder struct {
	mock *MockContestLogRepository
}

// NewMockContestLogRepository creates a new mock instance
func NewMockContestLogRepository(ctrl *gomock.Controller) *MockContestLogRepository {
	mock := &MockContestLogRepository{ctrl: ctrl}
	mock.recorder = &MockContestLogRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockContestLogRepository) EXPECT() *MockContestLogRepositoryMockRecorder {
	return m.recorder
}

// Store mocks base method
func (m *MockContestLogRepository) Store(contestLog *domain.ContestLog) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Store", contestLog)
	ret0, _ := ret[0].(error)
	return ret0
}

// Store indicates an expected call of Store
func (mr *MockContestLogRepositoryMockRecorder) Store(contestLog interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Store", reflect.TypeOf((*MockContestLogRepository)(nil).Store), contestLog)
}

// FindAll mocks base method
func (m *MockContestLogRepository) FindAll(contestID, userID uint64) (domain.ContestLogs, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindAll", contestID, userID)
	ret0, _ := ret[0].(domain.ContestLogs)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindAll indicates an expected call of FindAll
func (mr *MockContestLogRepositoryMockRecorder) FindAll(contestID, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindAll", reflect.TypeOf((*MockContestLogRepository)(nil).FindAll), contestID, userID)
}

// FindByID mocks base method
func (m *MockContestLogRepository) FindByID(id uint64) (domain.ContestLog, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindByID", id)
	ret0, _ := ret[0].(domain.ContestLog)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindByID indicates an expected call of FindByID
func (mr *MockContestLogRepositoryMockRecorder) FindByID(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByID", reflect.TypeOf((*MockContestLogRepository)(nil).FindByID), id)
}

// Delete mocks base method
func (m *MockContestLogRepository) Delete(id uint64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", id)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *MockContestLogRepositoryMockRecorder) Delete(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockContestLogRepository)(nil).Delete), id)
}

// FindRecent mocks base method
func (m *MockContestLogRepository) FindRecent(contestID, limit uint64) (domain.ContestLogs, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindRecent", contestID, limit)
	ret0, _ := ret[0].(domain.ContestLogs)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindRecent indicates an expected call of FindRecent
func (mr *MockContestLogRepositoryMockRecorder) FindRecent(contestID, limit interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindRecent", reflect.TypeOf((*MockContestLogRepository)(nil).FindRecent), contestID, limit)
}

// MockRankingRepository is a mock of RankingRepository interface
type MockRankingRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRankingRepositoryMockRecorder
}

// MockRankingRepositoryMockRecorder is the mock recorder for MockRankingRepository
type MockRankingRepositoryMockRecorder struct {
	mock *MockRankingRepository
}

// NewMockRankingRepository creates a new mock instance
func NewMockRankingRepository(ctrl *gomock.Controller) *MockRankingRepository {
	mock := &MockRankingRepository{ctrl: ctrl}
	mock.recorder = &MockRankingRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRankingRepository) EXPECT() *MockRankingRepositoryMockRecorder {
	return m.recorder
}

// Store mocks base method
func (m *MockRankingRepository) Store(contest domain.Ranking) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Store", contest)
	ret0, _ := ret[0].(error)
	return ret0
}

// Store indicates an expected call of Store
func (mr *MockRankingRepositoryMockRecorder) Store(contest interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Store", reflect.TypeOf((*MockRankingRepository)(nil).Store), contest)
}

// UpdateAmounts mocks base method
func (m *MockRankingRepository) UpdateAmounts(arg0 domain.Rankings) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateAmounts", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateAmounts indicates an expected call of UpdateAmounts
func (mr *MockRankingRepositoryMockRecorder) UpdateAmounts(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAmounts", reflect.TypeOf((*MockRankingRepository)(nil).UpdateAmounts), arg0)
}

// RankingsForContest mocks base method
func (m *MockRankingRepository) RankingsForContest(contestID uint64, languageCode domain.LanguageCode) (domain.Rankings, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RankingsForContest", contestID, languageCode)
	ret0, _ := ret[0].(domain.Rankings)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RankingsForContest indicates an expected call of RankingsForContest
func (mr *MockRankingRepositoryMockRecorder) RankingsForContest(contestID, languageCode interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RankingsForContest", reflect.TypeOf((*MockRankingRepository)(nil).RankingsForContest), contestID, languageCode)
}

// GlobalRankings mocks base method
func (m *MockRankingRepository) GlobalRankings(languageCode domain.LanguageCode) (domain.Rankings, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GlobalRankings", languageCode)
	ret0, _ := ret[0].(domain.Rankings)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GlobalRankings indicates an expected call of GlobalRankings
func (mr *MockRankingRepositoryMockRecorder) GlobalRankings(languageCode interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GlobalRankings", reflect.TypeOf((*MockRankingRepository)(nil).GlobalRankings), languageCode)
}

// FindAll mocks base method
func (m *MockRankingRepository) FindAll(contestID, userID uint64) (domain.Rankings, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindAll", contestID, userID)
	ret0, _ := ret[0].(domain.Rankings)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindAll indicates an expected call of FindAll
func (mr *MockRankingRepositoryMockRecorder) FindAll(contestID, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindAll", reflect.TypeOf((*MockRankingRepository)(nil).FindAll), contestID, userID)
}

// GetAllLanguagesForContestAndUser mocks base method
func (m *MockRankingRepository) GetAllLanguagesForContestAndUser(contestID, userID uint64) (domain.LanguageCodes, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllLanguagesForContestAndUser", contestID, userID)
	ret0, _ := ret[0].(domain.LanguageCodes)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllLanguagesForContestAndUser indicates an expected call of GetAllLanguagesForContestAndUser
func (mr *MockRankingRepositoryMockRecorder) GetAllLanguagesForContestAndUser(contestID, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllLanguagesForContestAndUser", reflect.TypeOf((*MockRankingRepository)(nil).GetAllLanguagesForContestAndUser), contestID, userID)
}

// CurrentRegistration mocks base method
func (m *MockRankingRepository) CurrentRegistration(userID uint64) (domain.RankingRegistration, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CurrentRegistration", userID)
	ret0, _ := ret[0].(domain.RankingRegistration)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CurrentRegistration indicates an expected call of CurrentRegistration
func (mr *MockRankingRepositoryMockRecorder) CurrentRegistration(userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CurrentRegistration", reflect.TypeOf((*MockRankingRepository)(nil).CurrentRegistration), userID)
}
