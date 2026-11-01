package contestwindow

import "time"

// Started takes a contest_start date at UTC midnight; a contest has started from the beginning of that UTC day.
func Started(start, now time.Time) bool {
	return !start.After(LatestStartedContestStart(now))
}

// Unfinished takes a contest_end date at UTC midnight; a contest accepts log changes through the end of that UTC day.
func Unfinished(end, now time.Time) bool {
	return !end.Before(EarliestUnfinishedContestEnd(now))
}

// RegistrationOpen takes a registration_end date at UTC midnight; a contest accepts new registrations through the end
// of that UTC day.
func RegistrationOpen(registrationEnd, now time.Time) bool {
	return !registrationEnd.Before(utcDate(now))
}

// LatestStartedContestStart is the latest contest_start a started contest can have; SQL selects started contests
// with contest_start <= it.
func LatestStartedContestStart(now time.Time) time.Time {
	return utcDate(now)
}

// EarliestUnfinishedContestEnd is the earliest contest_end an unfinished contest can have; SQL selects contests that
// accept log changes with contest_end >= it.
func EarliestUnfinishedContestEnd(now time.Time) time.Time {
	return utcDate(now)
}

func utcDate(t time.Time) time.Time {
	year, month, day := t.UTC().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
