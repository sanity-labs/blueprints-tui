package api

import (
	"encoding/json"
	"time"
)

type Stack struct {
	ID              string     `json:"id"`
	ScopeType       string     `json:"scopeType"`
	ScopeID         string     `json:"scopeId"`
	BlueprintID     string     `json:"blueprintId"`
	Name            string     `json:"name"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	RecentOperation *Operation `json:"recentOperation,omitempty"`
	Resources       []Resource `json:"resources,omitempty"`
	ResourceCount   *int       `json:"resourceCount,omitempty"`
}

// DisplayResourceCount prefers the full array length when available,
// falling back to the summary count from the list endpoint.
func (s Stack) DisplayResourceCount() *int {
	if len(s.Resources) > 0 {
		n := len(s.Resources)
		return &n
	}
	return s.ResourceCount
}

// Operation handles the API inconsistency where nested objects use snake_case
// but top-level responses use camelCase.
type Operation struct {
	ID          string     `json:"-"`
	StackID     string     `json:"-"`
	BlueprintID string     `json:"-"`
	Status      string     `json:"-"`
	CompletedAt *time.Time `json:"-"`
	CreatedAt   time.Time  `json:"-"`
	UpdatedAt   time.Time  `json:"-"`
}

func (o *Operation) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	decodeString := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := raw[k]; ok {
				var s string
				if json.Unmarshal(v, &s) == nil {
					return s
				}
			}
		}
		return ""
	}

	decodeTime := func(keys ...string) (time.Time, bool) {
		for _, k := range keys {
			if v, ok := raw[k]; ok {
				var t time.Time
				if json.Unmarshal(v, &t) == nil {
					return t, true
				}
			}
		}
		return time.Time{}, false
	}

	o.ID = decodeString("id")
	o.Status = decodeString("status")
	o.StackID = decodeString("stackId", "stack_id")
	o.BlueprintID = decodeString("blueprintId", "blueprint_id")

	if t, ok := decodeTime("createdAt", "created_at"); ok {
		o.CreatedAt = t
	}
	if t, ok := decodeTime("updatedAt", "updated_at"); ok {
		o.UpdatedAt = t
	}
	if t, ok := decodeTime("completedAt", "completed_at"); ok {
		o.CompletedAt = &t
	}

	return nil
}

type Resource struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	StackID          string         `json:"stackId"`
	OperationID      string         `json:"operationId"`
	BlueprintID      string         `json:"blueprintId,omitempty"`
	ExternalID       string         `json:"externalId,omitempty"`
	Type             string         `json:"type"`
	Parameters       map[string]any `json:"parameters"`
	ProviderMetadata map[string]any `json:"providerMetadata"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
}

type Log struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Level       string    `json:"level,omitempty"`
	Message     string    `json:"message"`
	Duration    int       `json:"duration,omitempty"`
	BlueprintID string    `json:"blueprintId,omitempty"`
	StackID     string    `json:"stackId,omitempty"`
	OperationID string    `json:"operationId,omitempty"`
	ResourceID  string    `json:"resourceId,omitempty"`
	RequestID   string    `json:"requestId,omitempty"`
}

type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Project struct {
	ID             string `json:"id"`
	DisplayName    string `json:"displayName"`
	OrganizationID string `json:"organizationId"`
}

type APIError struct {
	StatusCode int    `json:"statusCode"`
	Error      string `json:"error"`
	Message    string `json:"message"`
}
