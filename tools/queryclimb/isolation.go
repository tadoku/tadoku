package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

func isolation(ctx context.Context, root string, args []string) error {
	flags := flag.NewFlagSet("isolation", flag.ContinueOnError)
	baselineDB := flags.String("baseline-name", "", "owned baseline database")
	candidateDB := flags.String("candidate-name", "", "owned candidate database")
	candidateRoot := flags.String("candidate-root", "", "candidate checkout")
	target := flags.String("target", "", "one read-only SQL target")
	baselineCase := flags.String("baseline-buffer-case", "", "one baseline-only temporary-buffer case to certify")
	calibration := flags.String("calibration", "", "three comma-separated baseline observation files at this scale")
	scale := flags.Int("scale-index", 0, "zero-based scale position in the gate inputs")
	out := flags.String("output", "", "identity and observed-variance evidence JSON")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *candidateRoot == "" || *target == "" || *out == "" || *scale < 0 ||
		!ownedDatabase(*baselineDB) || !ownedDatabase(*candidateDB) || *baselineDB == *candidateDB {
		return errors.New("isolation needs candidate root, read-only target, output and valid scale")
	}
	if *baselineCase != "" && *calibration == "" {
		return errors.New("baseline buffer certification requires three calibration observations")
	}
	a, err := inventory(root)
	if err != nil {
		return err
	}
	b, err := inventory(*candidateRoot)
	if err != nil {
		return err
	}
	if len(a) != len(b) {
		return errors.New("SQL inventory differs")
	}
	proof := bufferProof{
		Target:            *target,
		BaselineDatabase:  *baselineDB,
		CandidateDatabase: *candidateDB,
		Queries:           map[string]string{},
		Variance:          map[string][]metrics{},
	}
	for _, original := range a {
		changed, err := findQuery(b, original.Key)
		if err != nil {
			return err
		}
		if original.Key == *target {
			if original.Write || changed.Write {
				return errors.New("identity exemption permits only a read-only target")
			}
			proof.TargetReadOnly = true
			if *baselineCase != "" {
				if original.SQL != changed.SQL {
					return errors.New("baseline buffer certification requires byte-identical target SQL")
				}
				proof.BaselineBuffer = &baselineBuffer{
					Key:     strconv.Itoa(*scale) + "/" + *target + "/" + *baselineCase,
					SQLHash: digest(original.SQL),
				}
			}
		} else if original.SQL != changed.SQL {
			return fmt.Errorf("non-target SQL differs: %s", original.Key)
		} else if !original.Write {
			proof.Queries[original.Key] = digest(original.SQL)
		}
	}
	if !proof.TargetReadOnly {
		return errors.New("target is missing")
	}
	before, err := generatedInterface(root)
	if err != nil {
		return err
	}
	after, err := generatedInterface(*candidateRoot)
	if err != nil {
		return err
	}
	if !sameJSON(before, after) {
		return errors.New("generated interface differs outside SQL strings")
	}

	proof.Baseline, err = databaseIdentity(ctx, root, *baselineDB)
	if err != nil {
		return err
	}
	proof.Candidate, err = databaseIdentity(ctx, *candidateRoot, *candidateDB)
	if err != nil {
		return err
	}
	if *calibration != "" {
		paths := strings.Split(*calibration, ",")
		if len(paths) != 3 {
			return errors.New("calibration requires three fresh-clone observations")
		}
		for _, path := range paths {
			var observations []metrics
			if err := readJSON(path, &observations); err != nil {
				return err
			}
			seen := map[string]bool{}
			for _, sample := range observations {
				baseline := proof.BaselineBuffer != nil && sample.Query == *target && sample.Case == *baselineCase
				if proof.Queries[sample.Query] == "" && !baseline {
					continue
				}
				key := strconv.Itoa(*scale) + "/" + sample.Query + "/" + sample.Case
				if seen[key] {
					return fmt.Errorf("duplicate calibration case: %s", key)
				}
				seen[key] = true
				sample.Plan, sample.Rows = nil, nil
				if !baseline {
					sample.EstimatedPlan = nil
				}
				proof.Variance[key] = append(proof.Variance[key], sample)
			}
		}
		for key, samples := range proof.Variance {
			if len(samples) != 3 {
				return fmt.Errorf("incomplete calibration: %s", key)
			}
			if samples[0].SharedBuffers == samples[1].SharedBuffers &&
				samples[1].SharedBuffers == samples[2].SharedBuffers &&
				samples[0].TempBuffers == samples[1].TempBuffers &&
				samples[1].TempBuffers == samples[2].TempBuffers {
				delete(proof.Variance, key)
			}
		}
	}
	if proof.BaselineBuffer != nil {
		cert := proof.BaselineBuffer
		identities := map[string]map[string]string{
			*baselineDB:  proof.Baseline,
			*candidateDB: proof.Candidate,
		}
		cert.Identities = map[string]map[string]string{}
		for _, sample := range proof.Variance[cert.Key] {
			if !ownedDatabase(sample.Database) {
				return errors.New("baseline witness must name an owned database")
			}
			identity, exists := identities[sample.Database]
			if !exists {
				identity, err = databaseIdentity(ctx, root, sample.Database)
				if err != nil {
					return err
				}
				identities[sample.Database] = identity
			}
			cert.Identities[sample.Database] = identity
		}
	}
	if err := writeJSON(*out, proof); err != nil {
		return err
	}
	if proof.BaselineBuffer != nil {
		key := proof.BaselineBuffer.Key
		var original, candidate metrics
		for _, sample := range proof.Variance[key] {
			if sample.Database == *baselineDB {
				original = sample
			}
			if sample.Database == *candidateDB {
				candidate = sample
			}
		}
		q := qualifications{proofs: map[string]bufferProof{key: proof}}
		if err := q.buffers(key, original, candidate, true); err != nil {
			return fmt.Errorf("baseline certification failed: %w", err)
		}
	}
	for _, part := range []string{"schema", "data", "statistics", "settings", "migrations"} {
		if proof.Baseline[part] != proof.Candidate[part] {
			return fmt.Errorf("isolation failed: %s differs", part)
		}
	}
	return output(map[string]any{"output": *out, "variable_cases": len(proof.Variance)})
}

func databaseIdentity(ctx context.Context, root, database string) (map[string]string, error) {
	conn, err := connect(ctx, localDSN, database, "tadoku_owner")
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)
	result := map[string]string{}
	for part, sql := range map[string]string{
		"schema":     schemaIdentitySQL,
		"statistics": statisticsIdentitySQL,
		"settings":   settingsIdentitySQL,
	} {
		var value string
		if err := conn.QueryRow(ctx, sql, pgx.QueryExecModeSimpleProtocol).Scan(&value); err != nil {
			return nil, fmt.Errorf("%s: %w", part, err)
		}
		result[part] = digest(value)
	}
	rows, err := conn.Query(ctx, `select c.relname
		from pg_class c
		join pg_namespace n on n.oid = c.relnamespace
		where n.nspname = 'public' and c.relkind in ('r', 'p')
		order by c.relname`)
	if err != nil {
		return nil, err
	}
	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			rows.Close()
			return nil, err
		}
		tables = append(tables, table)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	state, err := tableState(ctx, conn, tables, nil)
	if err != nil {
		return nil, err
	}
	var sequences string
	err = conn.QueryRow(ctx, `select coalesce(jsonb_agg(to_jsonb(s) order by sequencename), '[]')::text
		from pg_sequences s
		where schemaname = 'public'`).Scan(&sequences)
	if err != nil {
		return nil, err
	}
	result["data"] = digest(state + "/" + sequences)
	paths, err := filepath.Glob(filepath.Join(root, "services/tadoku-api/migrations/*.*.sql"))
	if err != nil {
		return nil, err
	}
	migrations := map[string]string{}
	for _, path := range paths {
		body, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		migrations[filepath.Base(path)] = digest(string(body))
	}
	encoded, err := json.Marshal(migrations)
	if err != nil {
		return nil, err
	}
	result["migrations"] = digest(string(encoded))
	return result, nil
}

const settingsIdentitySQL = `select jsonb_agg(jsonb_build_array(name, setting, unit) order by name)::text
	from pg_settings
	where category like 'Query Tuning%'
		or name in ('shared_buffers', 'work_mem', 'plan_cache_mode', 'jit', 'row_security',
			'search_path', 'max_parallel_workers', 'max_worker_processes', 'io_method',
			'max_parallel_workers_per_gather', 'parallel_leader_participation', 'hash_mem_multiplier',
			'effective_cache_size', 'effective_io_concurrency', 'server_version', 'lc_collate', 'lc_ctype',
			'TimeZone', 'DateStyle', 'extra_float_digits', 'bytea_output')`

const statisticsIdentitySQL = `select jsonb_build_object(
	'columns', (select jsonb_agg(to_jsonb(s) order by tablename, attname, inherited)
		from pg_stats s where schemaname = 'public'),
	'extended', (select jsonb_agg(to_jsonb(s) order by to_jsonb(s)::text)
		from pg_stats_ext s where schemaname = 'public'),
	'extended_expressions', (select jsonb_agg(to_jsonb(s) order by to_jsonb(s)::text)
		from pg_stats_ext_exprs s where schemaname = 'public'),
	'relations', (select jsonb_agg(jsonb_build_array(c.relname, c.relpages, c.reltuples, c.relallvisible)
		order by c.relname) from pg_class c join pg_namespace n on n.oid = c.relnamespace
		where n.nspname = 'public' and c.relkind in ('r', 'p', 'i')),
	'targets', (select jsonb_agg(jsonb_build_array(c.relname, a.attname, a.attstattarget)
		order by c.relname, a.attnum) from pg_class c join pg_namespace n on n.oid = c.relnamespace
		join pg_attribute a on a.attrelid = c.oid where n.nspname = 'public' and a.attnum > 0
		and not a.attisdropped)
)::text`

const schemaIdentitySQL = `select jsonb_build_object(
	'namespace', (select jsonb_build_array(n.nspowner::regrole::text, n.nspacl)
		from pg_namespace n where n.nspname = 'public'),
	'relations', (select jsonb_agg(jsonb_build_array(c.relname, c.relkind, c.relowner::regrole::text,
		c.relrowsecurity, c.relforcerowsecurity, c.relacl, c.reloptions) order by c.relname)
		from pg_class c join pg_namespace n on n.oid = c.relnamespace where n.nspname = 'public'),
	'columns', (select jsonb_agg(jsonb_build_array(c.relname, a.attnum, a.attname,
		format_type(a.atttypid, a.atttypmod), a.attnotnull, a.attidentity, a.attgenerated,
		a.attcollation::regcollation::text, pg_get_expr(d.adbin, d.adrelid)) order by c.relname, a.attnum)
		from pg_class c join pg_namespace n on n.oid = c.relnamespace
		join pg_attribute a on a.attrelid = c.oid
		left join pg_attrdef d on d.adrelid = c.oid and d.adnum = a.attnum
		where n.nspname = 'public' and a.attnum > 0 and not a.attisdropped),
	'constraints', (select jsonb_agg(jsonb_build_array(c.relname, co.conname,
		pg_get_constraintdef(co.oid), co.convalidated) order by c.relname, co.conname)
		from pg_constraint co join pg_class c on c.oid = co.conrelid
		join pg_namespace n on n.oid = c.relnamespace where n.nspname = 'public'),
	'indexes', (select jsonb_agg(to_jsonb(i) order by tablename, indexname)
		from pg_indexes i where schemaname = 'public'),
	'index_states', (select jsonb_agg(jsonb_build_array(c.relname, i.indisvalid, i.indisready,
		i.indislive, i.indisreplident, i.indisclustered) order by c.relname)
		from pg_index i join pg_class c on c.oid = i.indexrelid
		join pg_namespace n on n.oid = c.relnamespace where n.nspname = 'public'),
	'types', (select jsonb_agg(jsonb_build_array(t.typname, t.typtype, t.typnotnull,
		t.typbasetype::regtype::text, t.typtypmod, t.typcollation::regcollation::text,
		t.typdefault, t.typacl) order by t.typname)
		from pg_type t join pg_namespace n on n.oid = t.typnamespace where n.nspname = 'public'),
	'enum_values', (select jsonb_agg(jsonb_build_array(t.typname, e.enumlabel, e.enumsortorder)
		order by t.typname, e.enumsortorder) from pg_enum e join pg_type t on t.oid = e.enumtypid
		join pg_namespace n on n.oid = t.typnamespace where n.nspname = 'public'),
	'domain_constraints', (select jsonb_agg(jsonb_build_array(t.typname, c.conname,
		pg_get_constraintdef(c.oid), c.convalidated) order by t.typname, c.conname)
		from pg_constraint c join pg_type t on t.oid = c.contypid
		join pg_namespace n on n.oid = t.typnamespace where n.nspname = 'public'),
	'policies', (select jsonb_agg(to_jsonb(p) order by tablename, policyname)
		from pg_policies p where schemaname = 'public'),
	'functions', (select jsonb_agg(jsonb_build_array(p.oid::regprocedure::text,
		pg_get_functiondef(p.oid), p.proacl) order by p.oid::regprocedure::text)
		from pg_proc p join pg_namespace n on n.oid = p.pronamespace
		where n.nspname = 'public' and p.prokind in ('f', 'p')),
	'triggers', (select jsonb_agg(jsonb_build_array(c.relname, t.tgname, pg_get_triggerdef(t.oid))
		order by c.relname, t.tgname) from pg_trigger t join pg_class c on c.oid = t.tgrelid
		join pg_namespace n on n.oid = c.relnamespace where n.nspname = 'public'),
	'views', (select jsonb_agg(jsonb_build_array(c.relname, pg_get_viewdef(c.oid)) order by c.relname)
		from pg_class c join pg_namespace n on n.oid = c.relnamespace
		where n.nspname = 'public' and c.relkind in ('v', 'm'))
)::text`
