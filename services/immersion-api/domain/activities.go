package domain

import (
	"fmt"
	"sort"
)

var activities = []Activity{
	{ID: 1, Name: "Reading", Default: true},
	{ID: 2, Name: "Listening", Default: true},
	{ID: 3, Name: "Writing", Default: false},
	{ID: 4, Name: "Speaking", Default: false},
	{ID: 5, Name: "Study", Default: false},
}

var activitiesByID = map[int32]Activity{
	1: activities[0],
	2: activities[1],
	3: activities[2],
	4: activities[3],
	5: activities[4],
}

func Activities() []Activity {
	res := make([]Activity, len(activities))
	copy(res, activities)
	return res
}

func ActivityByID(id int32) (Activity, bool) {
	activity, ok := activitiesByID[id]
	return activity, ok
}

func ActivityName(id int) (string, bool) {
	activity, ok := ActivityByID(int32(id))
	return activity.Name, ok
}

func IsValidActivityID(id int32) bool {
	_, ok := ActivityByID(id)
	return ok
}

func ActivitiesByIDs(ids []int32) ([]Activity, bool) {
	res := make([]Activity, len(ids))
	for i, id := range ids {
		activity, ok := ActivityByID(id)
		if !ok {
			return nil, false
		}
		res[i] = activity
	}
	return res, true
}

func ActivitiesByIDsSortedByName(ids []int32) ([]Activity, bool) {
	res, ok := ActivitiesByIDs(ids)
	if !ok {
		return nil, false
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})
	return res, true
}

func requireActivity(id int32) (Activity, error) {
	activity, ok := ActivityByID(id)
	if !ok {
		return Activity{}, fmt.Errorf("activity %d is not valid: %w", id, ErrRequestInvalid)
	}
	return activity, nil
}

func hydrateActivity(activity *Activity) error {
	resolved, err := requireActivity(activity.ID)
	if err != nil {
		return err
	}
	*activity = resolved
	return nil
}

func hydrateActivities(activities []Activity) error {
	for i := range activities {
		if err := hydrateActivity(&activities[i]); err != nil {
			return err
		}
	}
	return nil
}

func hydrateActivitiesSortedByName(activities []Activity) error {
	if err := hydrateActivities(activities); err != nil {
		return err
	}
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].Name < activities[j].Name
	})
	return nil
}

func hydrateContestActivities(contest *ContestView, sortByName bool) error {
	if contest == nil {
		return nil
	}
	if sortByName {
		return hydrateActivitiesSortedByName(contest.AllowedActivities)
	}
	return hydrateActivities(contest.AllowedActivities)
}

func hydrateLogActivity(log *Log) error {
	activity, err := requireActivity(int32(log.ActivityID))
	if err != nil {
		return err
	}
	log.ActivityName = activity.Name
	return nil
}

func hydrateLogActivities(logs []Log) error {
	for i := range logs {
		if err := hydrateLogActivity(&logs[i]); err != nil {
			return err
		}
	}
	return nil
}

func hydrateActivityScores(scores []ActivityScore) error {
	for i := range scores {
		activity, err := requireActivity(int32(scores[i].ActivityID))
		if err != nil {
			return err
		}
		scores[i].ActivityName = activity.Name
	}
	return nil
}

func hydrateContestRegistrationActivities(registrations []ContestRegistration) error {
	for i := range registrations {
		if err := hydrateContestActivities(registrations[i].Contest, false); err != nil {
			return err
		}
	}
	return nil
}
