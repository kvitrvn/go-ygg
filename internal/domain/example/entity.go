package example

import "time"

// Example is a placeholder domain entity. Replace with your own.
type Example struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

func NewExample(id, name string) *Example {
	return &Example{
		ID:        id,
		Name:      name,
		CreatedAt: time.Now(),
	}
}
