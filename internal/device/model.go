package device

import "time"

type Device struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Brand     string    `json:"brand"`
	CreatedAt time.Time `json:"created_at"`
}
