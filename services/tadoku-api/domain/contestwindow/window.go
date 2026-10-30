package contestwindow

import "time"

// Started takes a contest_start date at UTC midnight; a contest has started from the beginning of that UTC day.
func Started(start, now time.Time) bool {
	return !start.After(StartedIfStartOnOrBefore(now))
}

// Open takes a contest_end date at UTC midnight; a contest accepts log changes through the end of that UTC day.
func Open(end, now time.Time) bool {
	return !end.Before(OpenIfEndOnOrAfter(now))
}

// StartedIfStartOnOrBefore is the bound SQL binds to select started contests: contest_start <= the bound.
func StartedIfStartOnOrBefore(now time.Time) time.Time {
	return utcDate(now)
}

// OpenIfEndOnOrAfter is the bound SQL binds to select contests that accept log changes: contest_end >= the bound.
func OpenIfEndOnOrAfter(now time.Time) time.Time {
	return utcDate(now)
}

func utcDate(t time.Time) time.Time {
	year, month, day := t.UTC().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
