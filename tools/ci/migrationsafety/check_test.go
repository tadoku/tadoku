package migrationsafety

import (
	"strings"
	"testing"
)

func TestCheck(t *testing.T) {
	tests := []struct {
		name string
		sql  string
		want []string
	}{
		{"blocking index", `create index users_email on users(email);`, []string{"create-index-without-concurrently"}},
		{"concurrent index", `create index concurrently users_email on users(email);`, nil},
		{"concurrent index in transaction", `begin; create index concurrently users_email on users(email); commit;`, []string{"concurrently-inside-transaction"}},
		{"concurrent index after transaction", `begin; select 1; commit; create index concurrently users_email on users(email);`, []string{"concurrently-inside-transaction"}},
		{"concurrent index with other statements", `create table x (id int); create index concurrently x_id on x(id);`, []string{"concurrently-inside-transaction"}},
		{"expression default", `alter table users add column token text default lower('X');`, []string{"add-column-expression-default"}},
		{"literal default", `alter table users add column enabled boolean default true;`, nil},
		{"type change", `alter table users alter column name type varchar(50);`, []string{"alter-column-type"}},
		{"validating foreign key", `alter table logs add constraint logs_user_fk foreign key (user_id) references users(id);`, []string{"foreign-key-validates-immediately"}},
		{"unvalidated foreign key", `alter table logs add constraint logs_user_fk foreign key (user_id) references users(id) not valid;`, nil},
		{"validate constraint", `alter table logs validate constraint logs_user_fk;`, nil},
		{"set not null", `alter table logs alter column user_id set not null;`, []string{"set-not-null"}},
		{"drop column", `alter table logs drop column old_value;`, []string{"drop-column"}},
		{"rename", `alter table logs rename to activity_logs;`, []string{"rename"}},
		{"unique constraint", `alter table users add constraint users_email_unique unique (email);`, []string{"unique-constraint"}},
		{"inline unique column", `alter table users add column email text unique;`, []string{"unique-constraint"}},
		{"inline foreign key column", `alter table logs add column user_id bigint references users(id);`, []string{"foreign-key-validates-immediately"}},
		{"unique using index", `alter table users add constraint users_email_unique unique using index users_email_idx;`, nil},
		{"truncate", `truncate table logs;`, []string{"truncate"}},
		{"vacuum full", `vacuum full logs;`, []string{"rewrite-command"}},
		{"cluster", `cluster logs using logs_id_idx;`, []string{"rewrite-command"}},
		{"multiple actions", `alter table logs add column token text default lower('X'), drop column old_value;`, []string{"add-column-expression-default", "drop-column"}},
		{"ordinary SQL", `create table users (id bigint primary key);`, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings, err := Check(tt.sql)
			if err != nil {
				t.Fatal(err)
			}
			if len(findings) != len(tt.want) {
				t.Fatalf("findings=%v, want rules %v", findings, tt.want)
			}
			for i, want := range tt.want {
				if findings[i].Rule != want {
					t.Errorf("finding %d rule=%q, want %q", i, findings[i].Rule, want)
				}
			}
		})
	}
}

func TestCheckReportsLineAndParseError(t *testing.T) {
	for sql, want := range map[string]int{
		"select 1;\ncreate index users_email on users(email);":                       2,
		"select 1;\n-- add an index\n\ncreate index users_email on users(email);":    4,
		"-- header\n/* block\ncomment */\ncreate index users_email on users(email);": 4,
	} {
		findings, err := Check(sql)
		if err != nil || len(findings) != 1 || findings[0].Line != want {
			t.Errorf("%q: findings=%v, err=%v; want one finding on line %d", sql, findings, err, want)
		}
	}
	if _, err := Check("alter table users add"); err == nil || !strings.Contains(err.Error(), "syntax error") {
		t.Fatalf("parse error=%v, want syntax error", err)
	}
}
