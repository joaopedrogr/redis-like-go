package command

// Type represents a command type
type Type string

const (
	// Command types
	SET     Type = "SET"
	GET     Type = "GET"
	DEL     Type = "DEL"
	EXPIRE  Type = "EXPIRE"
	TTL     Type = "TTL"
	PERSIST Type = "PERSIST"
	QUIT    Type = "QUIT"
)

// String returns the string representation of the command type
func (t Type) String() string {
	return string(t)
}

// IsValid checks if the command type is valid
func (t Type) IsValid() bool {
	switch t {
	case SET, GET, DEL, EXPIRE, TTL, PERSIST, QUIT:
		return true
	default:
		return false
	}
}

// IsWriteCommand checks if the command modifies data
func (t Type) IsWriteCommand() bool {
	switch t {
	case SET, DEL, EXPIRE, PERSIST:
		return true
	default:
		return false
	}
}
