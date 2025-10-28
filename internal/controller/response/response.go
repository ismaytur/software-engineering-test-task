package response

import "cruder/internal/model"

// User represents the user payload returned by controller endpoints.
type User = model.User

// Error wraps API error responses in a consistent schema.
type Error struct {
	Error string `json:"error"`
}
