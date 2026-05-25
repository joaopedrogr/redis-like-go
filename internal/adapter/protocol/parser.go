package protocol

import (
	"fmt"
	"strings"

	"redis-like-go/internal/domain/command"
)

// Command represents a parsed command
type Command struct {
	Type command.Type
	Args []string
}

// Parser handles command parsing and response formatting
type Parser struct{}

// NewParser creates a new protocol parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseCommand parses a command line into command and arguments
func (p *Parser) ParseCommand(line string) (*Command, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("empty command")
	}

	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	cmdType := command.Type(strings.ToUpper(parts[0]))
	if !cmdType.IsValid() {
		return nil, fmt.Errorf("unknown command: %s", parts[0])
	}

	cmd := &Command{
		Type: cmdType,
		Args: []string{},
	}

	if len(parts) > 1 {
		cmd.Args = parts[1:]
	}

	return cmd, nil
}

// FormatResponse formats a result into a response string
func (p *Parser) FormatResponse(result interface{}) string {
	switch v := result.(type) {
	case string:
		return v
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case bool:
		if v {
			return "OK"
		}
		return "ERR operation failed"
	case error:
		return fmt.Sprintf("ERR %s", v.Error())
	default:
		return fmt.Sprintf("%v", result)
	}
}

// FormatOK returns "OK" response
func (p *Parser) FormatOK() string {
	return "OK"
}

// FormatError returns an error response
func (p *Parser) FormatError(msg string) string {
	return fmt.Sprintf("ERR %s", msg)
}

// FormatNil returns "nil" response
func (p *Parser) FormatNil() string {
	return "nil"
}
