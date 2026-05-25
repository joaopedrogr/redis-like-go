package protocol

import (
	"testing"
)

func TestParseCommand(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		input    string
		wantCmd  string
		wantArgs []string
		wantErr  bool
	}{
		{
			name:     "SET command",
			input:    "SET key value",
			wantCmd:  "SET",
			wantArgs: []string{"key", "value"},
			wantErr:  false,
		},
		{
			name:     "GET command",
			input:    "GET key",
			wantCmd:  "GET",
			wantArgs: []string{"key"},
			wantErr:  false,
		},
		{
			name:     "SET with spaces in value",
			input:    "SET key hello world",
			wantCmd:  "SET",
			wantArgs: []string{"key", "hello", "world"},
			wantErr:  false,
		},
		{
			name:     "lowercase command",
			input:    "get key",
			wantCmd:  "GET",
			wantArgs: []string{"key"},
			wantErr:  false,
		},
		{
			name:     "command with extra spaces",
			input:    "  SET   key   value  ",
			wantCmd:  "SET",
			wantArgs: []string{"key", "value"},
			wantErr:  false,
		},
		{
			name:     "empty command",
			input:    "",
			wantCmd:  "",
			wantArgs: nil,
			wantErr:  true,
		},
		{
			name:     "command with only whitespace",
			input:    "   ",
			wantCmd:  "",
			wantArgs: nil,
			wantErr:  true,
		},
		{
			name:     "EXPIRE command",
			input:    "EXPIRE key 60",
			wantCmd:  "EXPIRE",
			wantArgs: []string{"key", "60"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := parser.ParseCommand(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if cmd.Type.String() != tt.wantCmd {
				t.Errorf("ParseCommand() cmd = %v, want %v", cmd.Type.String(), tt.wantCmd)
			}
			if len(cmd.Args) != len(tt.wantArgs) {
				t.Errorf("ParseCommand() args length = %v, want %v", len(cmd.Args), len(tt.wantArgs))
				return
			}
			for i := range cmd.Args {
				if cmd.Args[i] != tt.wantArgs[i] {
					t.Errorf("ParseCommand() args[%d] = %v, want %v", i, cmd.Args[i], tt.wantArgs[i])
				}
			}
		})
	}
}

func TestFormatResponse(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{
			name:  "string response",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "int response",
			input: 42,
			want:  "42",
		},
		{
			name:  "int64 response",
			input: int64(100),
			want:  "100",
		},
		{
			name:  "bool true",
			input: true,
			want:  "OK",
		},
		{
			name:  "bool false",
			input: false,
			want:  "ERR operation failed",
		},
		{
			name:  "error",
			input: "some error",
			want:  "some error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.FormatResponse(tt.input)
			if got != tt.want {
				t.Errorf("FormatResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatOK(t *testing.T) {
	parser := NewParser()
	got := parser.FormatOK()
	if got != "OK" {
		t.Errorf("FormatOK() = %v, want OK", got)
	}
}

func TestFormatError(t *testing.T) {
	parser := NewParser()
	got := parser.FormatError("test error")
	want := "ERR test error"
	if got != want {
		t.Errorf("FormatError() = %v, want %v", got, want)
	}
}

func TestFormatNil(t *testing.T) {
	parser := NewParser()
	got := parser.FormatNil()
	if got != "nil" {
		t.Errorf("FormatNil() = %v, want nil", got)
	}
}
