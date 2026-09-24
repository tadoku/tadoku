package asyncwork

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
)

type Type string

const (
	InvalidateContest  Type = "leaderboard.invalidate_contest.v1"
	InvalidateOfficial Type = "leaderboard.invalidate_official.v1"
)

func (t Type) Known() bool {
	return t == InvalidateContest || t == InvalidateOfficial
}

type Task struct {
	typ     Type
	payload json.RawMessage
}

func (t Task) Type() Type { return t.typ }

func (t Task) Payload() json.RawMessage { return bytes.Clone(t.payload) }

type ContestInvalidationPayload struct {
	ContestID uuid.UUID `json:"contest_id"`
}

type OfficialInvalidationPayload struct {
	Year int16 `json:"year"`
}

func NewContestInvalidation(id uuid.UUID) (Task, error) {
	payload := ContestInvalidationPayload{ContestID: id}
	if err := payload.Validate(); err != nil {
		return Task{}, err
	}
	encoded, err := json.Marshal(payload)
	return Task{typ: InvalidateContest, payload: encoded}, err
}

func NewOfficialInvalidation(year int16) (Task, error) {
	payload := OfficialInvalidationPayload{Year: year}
	if err := payload.Validate(); err != nil {
		return Task{}, err
	}
	encoded, err := json.Marshal(payload)
	return Task{typ: InvalidateOfficial, payload: encoded}, err
}

func (p ContestInvalidationPayload) Validate() error {
	if p.ContestID == uuid.Nil {
		return errors.New("contest_id must be a nonzero UUID")
	}
	return nil
}

func (p OfficialInvalidationPayload) Validate() error {
	if p.Year < 1 || p.Year > 9999 {
		return errors.New("year must be between 1 and 9999")
	}
	return nil
}

func DecodeContestInvalidation(raw json.RawMessage) (ContestInvalidationPayload, error) {
	var payload ContestInvalidationPayload
	if err := decode(raw, &payload); err != nil {
		return payload, err
	}
	return payload, payload.Validate()
}

func DecodeOfficialInvalidation(raw json.RawMessage) (OfficialInvalidationPayload, error) {
	var payload OfficialInvalidationPayload
	if err := decode(raw, &payload); err != nil {
		return payload, err
	}
	return payload, payload.Validate()
}

func decode(raw json.RawMessage, dst any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("decode task payload: %w", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return errors.New("task payload must contain one object")
	}
	return nil
}
