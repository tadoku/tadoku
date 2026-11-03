package e2e_test

import (
	"fmt"
	"strconv"
	"testing"
)

func shardValkeyDatabase(total, index string) (int, error) {
	databases := [...]int{15, 14, 12, 11, 10}
	if total == "" && index == "" {
		return databases[0], nil
	}

	count, err := strconv.Atoi(total)
	if err != nil || count < 0 || count > len(databases) {
		return 0, fmt.Errorf("invalid TEST_TOTAL_SHARDS %q: expected 0..%d", total, len(databases))
	}
	position, err := strconv.Atoi(index)
	if err != nil || position < 0 || position >= max(1, count) {
		return 0, fmt.Errorf("invalid TEST_SHARD_INDEX %q for TEST_TOTAL_SHARDS %q", index, total)
	}
	return databases[position], nil
}

func TestShardValkeyDatabase(t *testing.T) {
	for _, tc := range []struct {
		total string
		index string
		want  int
	}{
		{"", "", 15},
		{"0", "0", 15},
		{"1", "0", 15},
		{"5", "0", 15},
		{"5", "1", 14},
		{"5", "2", 12},
		{"5", "3", 11},
		{"5", "4", 10},
	} {
		t.Run(tc.total+"_"+tc.index, func(t *testing.T) {
			got, err := shardValkeyDatabase(tc.total, tc.index)
			if err != nil || got != tc.want {
				t.Errorf("shard (%q,%q): database=%d error=%v, want %d", tc.total, tc.index, got, err, tc.want)
			}
		})
	}

	for _, tc := range [][2]string{
		{"", "1"},
		{"0", "1"},
		{"1", "1"},
		{"2", ""},
		{"5", "5"},
		{"5", "-1"},
		{"5", "garbage"},
		{"garbage", "0"},
		{"-1", "0"},
		{"6", "0"},
	} {
		t.Run(tc[0]+"_"+tc[1], func(t *testing.T) {
			if _, err := shardValkeyDatabase(tc[0], tc[1]); err == nil {
				t.Errorf("accepted invalid shard (%q,%q)", tc[0], tc[1])
			}
		})
	}
}
