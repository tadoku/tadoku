package permissions

import (
	"testing"

	ketoclient "github.com/tadoku/tadoku/services/tadoku-api/infra/keto"
)

func TestAdmissionFromResultsMatchesByRelation(t *testing.T) {
	bannedCheck := ketoclient.PermissionCheck{Relation: "is_banned"}
	accessCheck := ketoclient.PermissionCheck{Relation: "access"}

	tests := []struct {
		name    string
		results []ketoclient.PermissionResult
		want    Admission
	}{
		{
			name: "banned when access precedes is_banned",
			results: []ketoclient.PermissionResult{
				{Check: accessCheck, Allowed: true},
				{Check: bannedCheck, Allowed: true},
			},
			want: Banned{},
		},
		{
			name: "admitted when access precedes is_banned",
			results: []ketoclient.PermissionResult{
				{Check: accessCheck, Allowed: true},
				{Check: bannedCheck, Allowed: false},
			},
			want: Admitted{},
		},
		{
			name: "no access when access precedes is_banned",
			results: []ketoclient.PermissionResult{
				{Check: accessCheck, Allowed: false},
				{Check: bannedCheck, Allowed: true},
			},
			want: NoAccess{},
		},
		{
			name: "banned in check order",
			results: []ketoclient.PermissionResult{
				{Check: bannedCheck, Allowed: true},
				{Check: accessCheck, Allowed: true},
			},
			want: Banned{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := admissionFromResults(test.results)
			if got != test.want {
				t.Errorf("admission=%T(%+v), want %T(%+v)", got, got, test.want, test.want)
			}
		})
	}
}
