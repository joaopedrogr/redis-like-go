package entity

// Item represents a key-value pair with optional expiration
type Item struct {
	Value     string
	ExpiresAt *int64 // Unix timestamp, nil if no expiration
}

// IsExpired checks if the item has expired based on current time
func (i *Item) IsExpired(now int64) bool {
	if i.ExpiresAt == nil {
		return false
	}
	return now >= *i.ExpiresAt
}
