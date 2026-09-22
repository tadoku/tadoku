package leaderboard

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/activities"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
	valkeygo "github.com/valkey-io/valkey-go"
)

const (
	contestPrefix = "leaderboard:contest:"
	yearlyPrefix  = "leaderboard:yearly:"
	globalKey     = "leaderboard:global"
)

var rebuildScript = valkeygo.NewLuaScript(`
redis.call('DEL', KEYS[1])
if #ARGV > 0 then
  return redis.call('ZADD', KEYS[1], unpack(ARGV))
end
return 0
`)

type Service struct {
	repository       *Repository
	client           valkeygo.Client
	operationTimeout time.Duration
}

func NewService(repository *Repository, client valkeygo.Client, operationTimeout time.Duration) *Service {
	return &Service{
		repository:       repository,
		client:           client,
		operationTimeout: operationTimeout,
	}
}

func (s *Service) FetchContest(ctx context.Context, request ContestRequest) (*Result, error) {
	request.Request = normalize(request.Request)
	if err := validateActivity(request.ActivityID); err != nil {
		return nil, err
	}
	if filtered(request.Request) {
		return s.fetchContestFromPostgres(ctx, request)
	}

	result, exists, err := s.fetchPage(ctx, contestPrefix+request.ContestID.String(), request.Page, request.PageSize)
	if err != nil {
		slog.WarnContext(ctx, "contest leaderboard cache unavailable; falling back to Postgres", "error", err)
		return s.fetchContestFromPostgres(ctx, request)
	}
	if !exists {
		scores, err := s.repository.allContestScores(ctx, request.ContestID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch all contest scores for rebuild: %w", err)
		}
		if err := s.rebuild(ctx, contestPrefix+request.ContestID.String(), scores); err != nil {
			slog.WarnContext(ctx, "contest leaderboard cache rebuild failed; falling back to Postgres", "error", err)
		}
		return s.fetchContestFromPostgres(ctx, request)
	}
	return cachedResult(result, request.Page, request.PageSize), nil
}

func (s *Service) FetchYearly(ctx context.Context, request YearlyRequest) (*Result, error) {
	request.Request = normalize(request.Request)
	if err := validateActivity(request.ActivityID); err != nil {
		return nil, err
	}
	if filtered(request.Request) {
		return postgresResult(s.repository.yearly(ctx, request))
	}

	key := yearlyPrefix + strconv.Itoa(int(request.Year))
	result, exists, err := s.fetchPage(ctx, key, request.Page, request.PageSize)
	if err != nil {
		slog.WarnContext(ctx, "yearly leaderboard cache unavailable; falling back to Postgres", "error", err)
		return postgresResult(s.repository.yearly(ctx, request))
	}
	if !exists {
		scores, err := s.repository.allYearlyScores(ctx, int(request.Year))
		if err != nil {
			return nil, fmt.Errorf("failed to fetch all yearly scores for rebuild: %w", err)
		}
		if err := s.rebuild(ctx, key, scores); err != nil {
			slog.WarnContext(ctx, "yearly leaderboard cache rebuild failed; falling back to Postgres", "error", err)
		}
		return postgresResult(s.repository.yearly(ctx, request))
	}
	return cachedResult(result, request.Page, request.PageSize), nil
}

func (s *Service) FetchGlobal(ctx context.Context, request Request) (*Result, error) {
	request = normalize(request)
	if err := validateActivity(request.ActivityID); err != nil {
		return nil, err
	}
	if filtered(request) {
		return postgresResult(s.repository.global(ctx, request))
	}

	result, exists, err := s.fetchPage(ctx, globalKey, request.Page, request.PageSize)
	if err != nil {
		slog.WarnContext(ctx, "global leaderboard cache unavailable; falling back to Postgres", "error", err)
		return postgresResult(s.repository.global(ctx, request))
	}
	if !exists {
		scores, err := s.repository.allGlobalScores(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch all global scores for rebuild: %w", err)
		}
		if err := s.rebuild(ctx, globalKey, scores); err != nil {
			slog.WarnContext(ctx, "global leaderboard cache rebuild failed; falling back to Postgres", "error", err)
		}
		return postgresResult(s.repository.global(ctx, request))
	}
	return cachedResult(result, request.Page, request.PageSize), nil
}

func (s *Service) fetchContestFromPostgres(ctx context.Context, request ContestRequest) (*Result, error) {
	exists, err := s.repository.contestExists(ctx, request.ContestID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errx.NewNotFoundError("contest not found")
	}
	return postgresResult(s.repository.contest(ctx, request))
}

func normalize(request Request) Request {
	if request.PageSize == 0 {
		request.PageSize = 25
	}
	if request.PageSize > 100 {
		request.PageSize = 100
	}
	return request
}

func validateActivity(activityID *int32) error {
	if activityID == nil {
		return nil
	}
	valid := false
	for _, activity := range activities.All() {
		if activity.ID == *activityID {
			valid = true
			break
		}
	}
	if !valid {
		return errx.NewInvalidInputError(fmt.Sprintf("activity %d is not valid", *activityID))
	}
	return nil
}

func filtered(request Request) bool { return request.LanguageCode != nil || request.ActivityID != nil }

func postgresResult(value *Leaderboard, err error) (*Result, error) {
	if err != nil {
		return nil, err
	}
	return &Result{Leaderboard: value}, nil
}

func cachedResult(cached *page, currentPage, pageSize int) *Result {
	return &Result{
		Leaderboard:         result(buildEntries(cached), cached.totalCount, currentPage, pageSize),
		HydrateDisplayNames: true,
	}
}

func buildEntries(cached *page) []Entry {
	if len(cached.scores) == 0 {
		return []Entry{}
	}
	entries := make([]Entry, len(cached.scores))
	rank := cached.startRank
	for i, item := range cached.scores {
		if i > 0 && item.value < cached.scores[i-1].value {
			rank = cached.startRank + i
		}
		entries[i] = Entry{
			Rank:   rank,
			UserID: item.userID,
			Score:  float32(item.value),
		}
	}
	for i := range entries {
		if i > 0 && entries[i].Rank == entries[i-1].Rank {
			entries[i].IsTie = true
			entries[i-1].IsTie = true
		}
	}
	if cached.hasPrevTie {
		entries[0].IsTie = true
	}
	if cached.hasNextTie {
		entries[len(entries)-1].IsTie = true
	}
	return entries
}

func (s *Service) fetchPage(ctx context.Context, key string, currentPage, pageSize int) (*page, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, s.operationTimeout)
	defer cancel()
	exists, err := s.client.Do(ctx, s.client.B().Exists().Key(key+":last_updated").Build()).AsInt64()
	if err != nil {
		return nil, false, fmt.Errorf("check leaderboard marker %s: %w", key, err)
	}
	if exists == 0 {
		return nil, false, nil
	}
	total, err := s.client.Do(ctx, s.client.B().Zcard().Key(key).Build()).AsInt64()
	if err != nil {
		return nil, false, fmt.Errorf("get leaderboard cardinality %s: %w", key, err)
	}
	if total == 0 {
		return &page{scores: []score{}, startRank: 1}, true, nil
	}

	start := int64(currentPage * pageSize)
	stop := start + int64(pageSize) - 1
	fetchStart := start
	if fetchStart > 0 {
		fetchStart--
	}
	values, err := s.client.Do(ctx, s.client.B().Zrange().Key(key).Min(strconv.FormatInt(fetchStart, 10)).Max(strconv.FormatInt(stop+1, 10)).Rev().Withscores().Build()).AsZScores()
	if err != nil {
		return nil, false, fmt.Errorf("fetch leaderboard page %s: %w", key, err)
	}
	hasPrevious := fetchStart < start && len(values) > 0
	previous := float64(0)
	if hasPrevious {
		previous = values[0].Score
		values = values[1:]
	}
	hasNext := int64(len(values)) > int64(pageSize)
	next := float64(0)
	if hasNext {
		next = values[len(values)-1].Score
		values = values[:len(values)-1]
	}
	scores := make([]score, len(values))
	for i, value := range values {
		id, err := uuid.Parse(value.Member)
		if err != nil {
			return nil, false, fmt.Errorf("parse leaderboard member %q: %w", value.Member, err)
		}
		scores[i] = score{userID: id, value: value.Score}
	}
	startRank := int(start) + 1
	if len(scores) > 0 {
		higher, err := s.client.Do(ctx, s.client.B().Zcount().Key(key).Min("("+strconv.FormatFloat(scores[0].value, 'f', -1, 64)).Max("+inf").Build()).AsInt64()
		if err != nil {
			return nil, false, fmt.Errorf("count higher leaderboard scores %s: %w", key, err)
		}
		startRank = int(higher) + 1
	}
	return &page{
		scores:     scores,
		totalCount: int(total),
		startRank:  startRank,
		hasPrevTie: hasPrevious && len(scores) > 0 && previous == scores[0].value,
		hasNextTie: hasNext && len(scores) > 0 && next == scores[len(scores)-1].value,
	}, true, nil
}

func (s *Service) rebuild(ctx context.Context, key string, scores []score) error {
	ctx, cancel := context.WithTimeout(ctx, s.operationTimeout)
	defer cancel()
	args := make([]string, 0, len(scores)*2)
	for _, item := range scores {
		args = append(args, strconv.FormatFloat(item.value, 'f', -1, 64), item.userID.String())
	}
	if err := rebuildScript.Exec(ctx, s.client, []string{key}, args).Error(); err != nil {
		return fmt.Errorf("rebuild leaderboard %s: %w", key, err)
	}
	if err := s.client.Do(ctx, s.client.B().Set().Key(key+":last_updated").Value(strconv.FormatInt(timex.Now().Unix(), 10)).Build()).Error(); err != nil {
		return fmt.Errorf("set leaderboard marker %s: %w", key, err)
	}
	return nil
}
