package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type pageOracle struct {
	Query      string      `json:"query"`
	Case       string      `json:"case"`
	Tenant     string      `json:"tenant"`
	Args       []*string   `json:"args"`
	TieColumns []int       `json:"tie_columns"`
	OffsetArg  int         `json:"offset_arg"`
	LimitArg   int         `json:"limit_arg"`
	SQLHash    string      `json:"sql_hash"`
	Database   string      `json:"database"`
	Rows       [][]*string `json:"rows"`
}

type baselineBuffer struct {
	Key        string                       `json:"key"`
	SQLHash    string                       `json:"sql_hash"`
	Identities map[string]map[string]string `json:"identities"`
}

type bufferProof struct {
	BaselineBuffer    *baselineBuffer      `json:"baseline_buffer,omitempty"`
	Target            string               `json:"target"`
	TargetReadOnly    bool                 `json:"target_read_only"`
	BaselineDatabase  string               `json:"baseline_database"`
	CandidateDatabase string               `json:"candidate_database"`
	Baseline          map[string]string    `json:"baseline"`
	Candidate         map[string]string    `json:"candidate"`
	Queries           map[string]string    `json:"queries"`
	Variance          map[string][]metrics `json:"variance"`
}

type qualifications struct {
	pages  map[string]pageOracle
	proofs map[string]bufferProof
}

func readQualifications(pageFiles, proofFiles, target string, determinism bool) (qualifications, error) {
	q := qualifications{pages: map[string]pageOracle{}, proofs: map[string]bufferProof{}}
	if pageFiles != "" {
		for scale, path := range strings.Split(pageFiles, ",") {
			var pages []pageOracle
			if err := readJSON(path, &pages); err != nil {
				return q, err
			}
			for _, page := range pages {
				key := fmt.Sprintf("%d/%s/%s", scale, page.Query, page.Case)
				if _, exists := q.pages[key]; exists {
					return q, fmt.Errorf("duplicate page oracle: %s", key)
				}
				q.pages[key] = page
			}
		}
	}
	if proofFiles == "" {
		return q, nil
	}
	for _, path := range strings.Split(proofFiles, ",") {
		var proof bufferProof
		if err := readJSON(path, &proof); err != nil {
			return q, err
		}
		if target == "" || proof.Target != target || !proof.TargetReadOnly {
			return q, errors.New("buffer proof requires the same read-only target")
		}
		if !ownedDatabase(proof.BaselineDatabase) || !ownedDatabase(proof.CandidateDatabase) ||
			proof.BaselineDatabase == proof.CandidateDatabase {
			return q, errors.New("buffer proof needs two distinct owned measurement databases")
		}
		for _, part := range []string{"schema", "data", "statistics", "settings", "migrations"} {
			if !isDigest(proof.Baseline[part]) || proof.Baseline[part] != proof.Candidate[part] {
				return q, fmt.Errorf("buffer proof differs or omits %s", part)
			}
		}
		if proof.BaselineBuffer != nil && !determinism {
			return q, errors.New("baseline buffer certification requires determinism mode")
		}
		for key := range proof.Variance {
			if _, exists := q.proofs[key]; exists {
				return q, fmt.Errorf("duplicate buffer proof: %s", key)
			}
			q.proofs[key] = proof
		}
	}
	return q, nil
}

func isDigest(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 32
}

func sameJSON(a, b any) bool {
	x, errA := json.Marshal(a)
	y, errB := json.Marshal(b)
	return errA == nil && errB == nil && string(x) == string(y)
}

func samePlan(a, b json.RawMessage) bool {
	var left, right any
	if len(a) == 0 || len(b) == 0 || json.Unmarshal(a, &left) != nil || json.Unmarshal(b, &right) != nil {
		return false
	}
	return left != nil && right != nil && sameJSON(left, right)
}

func (q qualifications) buffers(key string, original, candidate metrics, determinism bool) error {
	proof, exists := q.proofs[key]
	if !exists {
		return errors.New("no documented buffer-variance proof")
	}
	if original.Database != proof.BaselineDatabase || candidate.Database != proof.CandidateDatabase {
		return errors.New("buffer identity proof belongs to different measurement databases")
	}
	baseline := original.Query == proof.Target
	expected := proof.Queries[original.Query]
	if baseline {
		cert := proof.BaselineBuffer
		if !determinism || cert == nil || cert.Key != key {
			return errors.New("no scoped baseline-only buffer certification")
		}
		expected = cert.SQLHash
		if err := proof.baselineIdentities(); err != nil {
			return err
		}
		if !sameBaselineWork(original, candidate) {
			return errors.New("baseline observations differ outside temporary buffers")
		}
	}
	if !isDigest(expected) ||
		original.SQLHash != expected || candidate.SQLHash != expected {
		return errors.New("buffer proof does not cover byte-identical unchanged SQL")
	}
	if original.Tenant != candidate.Tenant || !sameJSON(original.Args, candidate.Args) ||
		!sameJSON(original.WarmArgs, candidate.WarmArgs) || !sameJSON(original.TieColumns, candidate.TieColumns) ||
		!sameJSON(original.MaskColumns, candidate.MaskColumns) {
		return errors.New("unchanged-query case inputs differ")
	}
	if original.PlanHash != candidate.PlanHash || !samePlan(original.EstimatedPlan, candidate.EstimatedPlan) ||
		original.GenericPlans != candidate.GenericPlans || original.CustomPlans != candidate.CustomPlans {
		return errors.New("full estimated plans or prepared-plan counters differ")
	}
	witness := proof.Variance[key]
	if len(witness) != 3 {
		return errors.New("buffer proof needs three baseline observations")
	}
	varied := false
	observedDatabases := map[string]bool{}
	for _, observed := range witness {
		if !ownedDatabase(observed.Database) || observedDatabases[observed.Database] {
			return errors.New("buffer variance needs three distinct owned fresh clones")
		}
		observedDatabases[observed.Database] = true
		first := witness[0]
		if baseline && !sameBaselineWork(original, observed) {
			return errors.New("baseline witnesses differ outside temporary buffers")
		}
		if observed.Query != original.Query || observed.Case != original.Case ||
			observed.SQLHash != expected || observed.PlanHash != original.PlanHash ||
			observed.Fingerprint != original.Fingerprint || observed.RowsProcessed != original.RowsProcessed ||
			observed.GenericPlans != original.GenericPlans || observed.CustomPlans != original.CustomPlans ||
			!sameJSON(observed.Args, original.Args) || !sameJSON(observed.WarmArgs, original.WarmArgs) ||
			observed.Tenant != original.Tenant {
			return errors.New("variance observations differ outside buffer counters")
		}
		varied = varied || observed.SharedBuffers != first.SharedBuffers || observed.TempBuffers != first.TempBuffers
	}
	if !varied {
		return errors.New("baseline buffers did not vary in the supplied observations")
	}
	return nil
}

func (proof bufferProof) baselineIdentities() error {
	cert := proof.BaselineBuffer
	witness := proof.Variance[cert.Key]
	if len(witness) != 3 || len(cert.Identities) != 3 || len(proof.Baseline) != 5 {
		return errors.New("baseline certification needs three complete clone identities and witnesses")
	}
	for _, part := range []string{"schema", "data", "statistics", "settings", "migrations"} {
		if !isDigest(proof.Baseline[part]) {
			return fmt.Errorf("baseline certification omits %s", part)
		}
	}
	if !sameJSON(proof.Baseline, proof.Candidate) ||
		!sameJSON(cert.Identities[proof.BaselineDatabase], proof.Baseline) ||
		!sameJSON(cert.Identities[proof.CandidateDatabase], proof.Candidate) {
		return errors.New("baseline certification identities differ from the measured databases")
	}
	for _, sample := range witness {
		if !sameJSON(cert.Identities[sample.Database], proof.Baseline) {
			return errors.New("baseline witness identity is missing or differs")
		}
	}
	return nil
}

func sameBaselineWork(a, b metrics) bool {
	if !samePlan(a.EstimatedPlan, b.EstimatedPlan) || a.GenericPlans+a.CustomPlans != 6 {
		return false
	}
	a.Plan, b.Plan = nil, nil
	a.EstimatedPlan, b.EstimatedPlan = nil, nil
	a.TempBuffers, b.TempBuffers = 0, 0
	a.Database, b.Database = "", ""
	return sameJSON(a, b)
}

func (q qualifications) page(key string, original, candidate metrics) error {
	oracle, exists := q.pages[key]
	if !exists {
		return errors.New("no baseline legal-page oracle")
	}
	if !ownedDatabase(oracle.Database) || oracle.Database != original.Database ||
		!ownedDatabase(candidate.Database) ||
		oracle.Query != original.Query || oracle.Case != original.Case || oracle.Tenant != original.Tenant ||
		!sameJSON(oracle.Args, original.Args) || !sameJSON(oracle.Args, candidate.Args) ||
		candidate.Tenant != original.Tenant || !isDigest(oracle.SQLHash) || oracle.SQLHash != original.SQLHash ||
		len(oracle.TieColumns) == 0 || !sameJSON(oracle.TieColumns, original.TieColumns) ||
		!sameJSON(oracle.TieColumns, candidate.TieColumns) {
		return errors.New("page oracle is not bound to the baseline query and exact case")
	}
	if oracle.OffsetArg < 0 || oracle.LimitArg < 0 || oracle.OffsetArg == oracle.LimitArg ||
		oracle.OffsetArg >= len(oracle.Args) || oracle.LimitArg >= len(oracle.Args) ||
		oracle.Args[oracle.OffsetArg] == nil || oracle.Args[oracle.LimitArg] == nil {
		return errors.New("invalid pagination argument positions")
	}
	offset, err := strconv.Atoi(*oracle.Args[oracle.OffsetArg])
	if err != nil || offset < 0 || offset >= len(oracle.Rows) {
		return errors.New("empty or invalid pages require the exact fingerprint check")
	}
	limit, err := strconv.Atoi(*oracle.Args[oracle.LimitArg])
	if err != nil || limit <= 0 {
		return errors.New("invalid page size")
	}
	end := offset + min(limit, len(oracle.Rows)-offset)
	for _, actual := range []metrics{original, candidate} {
		if actual.Rows == nil || len(actual.Rows) != actual.Returned || len(actual.MaskColumns) != 0 {
			return errors.New("page has no complete returned-row evidence")
		}
		fp, err := fingerprint(actual.Rows, actual.TieColumns, nil)
		if err != nil || fp != actual.Fingerprint {
			return errors.New("row evidence does not match the measured fingerprint")
		}
		if err := legalPage(oracle.Rows, actual.Rows, oracle.TieColumns, offset, end); err != nil {
			return err
		}
	}
	return nil
}

func legalPage(universe, actual [][]*string, columns []int, start, end int) error {
	type group struct {
		rows map[string]int
		need int
	}
	groups := []group{}
	positions := map[string]int{}
	keys := []string{}
	keyOf := func(row []*string) (string, error) {
		key := []*string{}
		for _, col := range columns {
			if col < 0 || col >= len(row) {
				return "", errors.New("ordering column out of range")
			}
			key = append(key, row[col])
		}
		encoded, _ := json.Marshal(key)
		return string(encoded), nil
	}

	split := false
	for i := 0; i < len(universe); {
		key, err := keyOf(universe[i])
		if err != nil {
			return err
		}
		if _, seen := positions[key]; seen {
			return errors.New("oracle ordering groups are not contiguous")
		}
		j := i
		rows := map[string]int{}
		for j < len(universe) {
			next, err := keyOf(universe[j])
			if err != nil {
				return err
			}
			if next != key {
				break
			}
			row, _ := json.Marshal(universe[j])
			rows[string(row)]++
			j++
		}
		split = split || i < start && start < j || i < end && end < j
		positions[key] = len(groups)
		keys = append(keys, key)
		groups = append(groups, group{rows: rows, need: max(0, min(j, end)-max(i, start))})
		i = j
	}
	if !split || len(actual) != end-start {
		return errors.New("not a complete boundary-cut page")
	}

	previous := -1
	for _, row := range actual {
		key, err := keyOf(row)
		if err != nil {
			return err
		}
		position, exists := positions[key]
		if !exists || position < previous {
			return errors.New("page contains an unknown or out-of-order group")
		}
		encoded, _ := json.Marshal(row)
		g := &groups[position]
		if g.need == 0 || g.rows[string(encoded)] == 0 {
			return errors.New("page contains an illegal row value, multiplicity or boundary cardinality")
		}
		g.rows[string(encoded)]--
		g.need--
		previous = position
	}
	for i, g := range groups {
		if g.need != 0 {
			return fmt.Errorf("page omits required rows from ordering group %s", keys[i])
		}
	}
	return nil
}
