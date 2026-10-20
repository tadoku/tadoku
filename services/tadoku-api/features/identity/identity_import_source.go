package identity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/mail"
	"golang.org/x/crypto/bcrypt"
)

type sourceIdentity struct {
	id                     uuid.UUID
	email, normalizedEmail string
	displayName, hash      string
	state                  string
	createdAt, updatedAt   time.Time
}

var bcryptFormat = regexp.MustCompile(`^\$2[aby]\$[0-9]{2}\$[./A-Za-z0-9]{53}$`)

func readSource(ctx context.Context, pool *pgxpool.Pool, expect int, output io.Writer) ([]sourceIdentity, error) {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, databaseError("source snapshot", err)
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(cleanup)
	}()

	var networks, identities, passwords, credentials int
	err = tx.QueryRow(ctx, `select
		(select count(*) from networks),
		(select count(*) from identities),
		(select count(*)
		 from identity_credentials c
		 join identity_credential_types ct on ct.id = c.identity_credential_type_id
		 where ct.name = 'password'),
		(select count(*) from identity_credentials)`).Scan(&networks, &identities, &passwords, &credentials)
	if err != nil {
		return nil, databaseError("source read", err)
	}
	_, err = fmt.Fprintf(
		output,
		"source_networks=%d source_identities=%d password_credentials=%d\n",
		networks,
		identities,
		passwords,
	)
	if err != nil {
		return nil, err
	}
	if networks != 1 {
		return nil, errors.New("preflight network: require exactly one source network")
	}
	if identities == 0 || (expect > 0 && identities != expect) {
		return nil, errors.New("preflight source_count: empty source or --expect-count mismatch")
	}
	if passwords != identities {
		return nil, errors.New("preflight password_credentials: require one password credential per identity")
	}
	if credentials != passwords {
		return nil, errors.New("preflight credential_types: non-password credentials require separate review")
	}

	rows, err := tx.Query(ctx, `select i.id, i.traits, i.state, i.created_at, i.updated_at, c.config, i.schema_id
		from identities i
		left join identity_credentials c on c.identity_id = i.id
		  and c.nid = i.nid
		  and c.identity_credential_type_id = (select id from identity_credential_types where name = 'password')
		where i.nid = (select id from networks)
		order by i.id`)
	if err != nil {
		return nil, databaseError("source read", err)
	}
	defer rows.Close()

	var result []sourceIdentity
	var incompatible []uuid.UUID
	seenIDs := make(map[uuid.UUID]bool)
	seenEmails := make(map[string]bool)
	for rows.Next() {
		var record sourceIdentity
		var traits, config []byte
		var schema string
		err := rows.Scan(&record.id, &traits, &record.state, &record.createdAt, &record.updatedAt, &config, &schema)
		if err != nil {
			return nil, databaseError("source read", err)
		}
		if schema != "user" {
			return nil, fmt.Errorf("preflight user_id=%s field=schema_id", record.id)
		}
		if record.id == uuid.Nil || seenIDs[record.id] || config == nil {
			return nil, errors.New("preflight password_credentials: invalid UUID or credential cardinality")
		}
		seenIDs[record.id] = true

		var rawTraits struct {
			Email       string `json:"email"`
			DisplayName string `json:"display_name"`
		}
		err = json.Unmarshal(traits, &rawTraits)
		if err != nil {
			return nil, fmt.Errorf("preflight user_id=%s field=traits", record.id)
		}
		email, err := ParseEmail(rawTraits.Email)
		if err != nil {
			return nil, fmt.Errorf("preflight user_id=%s field=email", record.id)
		}
		if _, err := ParseDisplayName(rawTraits.DisplayName); err != nil {
			return nil, fmt.Errorf("preflight user_id=%s field=display_name", record.id)
		}
		record.email = rawTraits.Email
		record.normalizedEmail = email.String()
		record.displayName = rawTraits.DisplayName
		if seenEmails[record.normalizedEmail] {
			return nil, fmt.Errorf("preflight user_id=%s field=email_normalized duplicate", record.id)
		}
		seenEmails[record.normalizedEmail] = true
		if !mail.ValidAddress(record.email) {
			incompatible = append(incompatible, record.id)
		}

		var password struct {
			Hash          string `json:"hashed_password"`
			MigrationHook bool   `json:"use_password_migration_hook"`
		}
		err = json.Unmarshal(config, &password)
		if err != nil || password.MigrationHook || !bcryptFormat.MatchString(password.Hash) {
			return nil, fmt.Errorf("preflight user_id=%s field=password_hash invalid bcrypt config", record.id)
		}
		if _, err := bcrypt.Cost([]byte(password.Hash)); err != nil {
			return nil, fmt.Errorf("preflight user_id=%s field=password_hash invalid bcrypt cost", record.id)
		}
		record.hash = password.Hash
		if record.createdAt.IsZero() || record.updatedAt.IsZero() ||
			(record.state != "active" && record.state != "inactive") {
			return nil, fmt.Errorf("preflight user_id=%s field=state_or_timestamps", record.id)
		}
		result = append(result, record)
	}
	if err := rows.Err(); err != nil {
		return nil, databaseError("source read", err)
	}
	if len(result) != identities {
		return nil, errors.New("preflight password_credentials: source snapshot count mismatch")
	}
	for _, id := range incompatible {
		_, err := fmt.Fprintf(output, "preflight user_id=%s field=email reason=recipient_rule\n", id)
		if err != nil {
			return nil, err
		}
	}
	if len(incompatible) > 0 {
		return nil, fmt.Errorf("preflight email: %d stored addresses fail the email recipient rule", len(incompatible))
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, databaseError("source snapshot", err)
	}
	return result, nil
}
