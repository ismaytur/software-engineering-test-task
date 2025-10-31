package model

import "time"

type APIKey struct {
	ID         int       `json:"id"`
	KeyHash    string    `json:"-"`
	ClientName string    `json:"client_name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
