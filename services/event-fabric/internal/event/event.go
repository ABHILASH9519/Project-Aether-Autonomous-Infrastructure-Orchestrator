package event

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type Envelope struct {
	ID string `json:"event_id"`
	Type string `json:"event_type"`
	Version int `json:"schema_version"`
	OccurredAt time.Time `json:"occurred_at"`
	Source string `json:"source"`
	CorrelationID string `json:"correlation_id"`
	Payload json.RawMessage `json:"payload"`
}

func (e Envelope) Validate() error {
	if strings.TrimSpace(e.ID)=="" || len(strings.TrimSpace(e.Type))<3 || e.Version!=1 || e.OccurredAt.IsZero() || strings.TrimSpace(e.Source)=="" || strings.TrimSpace(e.CorrelationID)=="" {
		return errors.New("invalid event metadata")
	}
	var payload map[string]json.RawMessage
	if json.Unmarshal(e.Payload,&payload)!=nil || payload==nil {
		return errors.New("payload must be a JSON object")
	}
	return nil
}
