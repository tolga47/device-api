package device

import "time"

// Device is a single device record. ID and CreatedAt are set by the server.
type Device struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Brand     string    `json:"brand"`
	CreatedAt time.Time `json:"created_at"`
}
