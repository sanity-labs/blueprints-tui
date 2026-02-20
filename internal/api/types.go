package api

import "time"

type Stack struct {
	ID          string    `json:"id"`
	ScopeType   string    `json:"scopeType"`
	ScopeID     string    `json:"scopeId"`
	BlueprintID string    `json:"blueprintId"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Operation struct {
	ID          string     `json:"id"`
	StackID     string     `json:"stackId"`
	BlueprintID string     `json:"blueprintId"`
	Status      string     `json:"status"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type Resource struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	StackID          string                 `json:"stackId"`
	OperationID      string                 `json:"operationId"`
	BlueprintID      string                 `json:"blueprintId,omitempty"`
	ExternalID       string                 `json:"externalId,omitempty"`
	Type             string                 `json:"type"`
	Parameters       map[string]any `json:"parameters"`
	ProviderMetadata map[string]any `json:"providerMetadata"`
	CreatedAt        time.Time              `json:"createdAt"`
	UpdatedAt        time.Time              `json:"updatedAt"`
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

type APIError struct {
	StatusCode int    `json:"statusCode"`
	Error      string `json:"error"`
	Message    string `json:"message"`
}
