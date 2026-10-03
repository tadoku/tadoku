package main

import (
	"context"
	"testing"
)

func TestCommandRefusalsBeforeConfiguration(t *testing.T) {
	for _, args := range [][]string{
		{}, {"unknown"}, {"provision", "--tenant", "tadoku/prod", "--flipt-features", "/absent"},
		{"teardown", "--tenant", "tadoku"}, {"teardown", "--tenant", "branch-0123abcd"},
		{"teardown", "--tenant", "tadoku/branch"}, {"teardown", "--tenant", "tadoku/branch-0123abcd", "extra"},
		{"provision", "--tenant", "tadoku/branch-0123abcd"},
		{"provision", "--tenant", "tadoku/branch-0123abcd", "--flipt-features", "/absent", "--tester", "invalid"},
		{"provision", "--tenant", "tadoku/branch-0123abcd", "--flipt-features", "/absent",
			"--tester", "00000000-0000-0000-0000-000000000000"},
		{"override", "set", "--tenant", "tadoku/prod", "--component", "tadoku-worker"},
		{"override", "set", "--tenant", "tadoku/branch-0123abcd", "--component", "unregistered-worker"},
		{"override", "clear", "--tenant", "tadoku/branch-0123abcd", "--component", "unregistered-worker"},
	} {
		if _, err := parseCommand(args); err == nil {
			t.Errorf("accepted unsafe command %v", args)
		}
		if err := run(context.Background(), args); err == nil {
			t.Errorf("unsafe command ran: %v", args)
		}
	}
}

func TestCommandParsesRepeatedTesters(t *testing.T) {
	command, err := parseCommand([]string{
		"provision", "--tenant", "tadoku/branch-0123abcd", "--flipt-features", "features.yaml",
		"--tester", "11111111-1111-4111-8111-111111111111",
		"--tester", "22222222-2222-4222-8222-222222222222",
	})
	if err != nil || len(command.testers) != 2 || command.key.String() != "tadoku/branch-0123abcd" {
		t.Fatalf("parsed command=%+v error=%v", command, err)
	}
}
