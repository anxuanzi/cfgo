package cfgo

import (
	"strings"
	"testing"
)

func TestParseEnvLine(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantKey   string
		wantValue string
		wantOK    bool
	}{
		{"basic pair", "KEY=value", "KEY", "value", true},
		{"spaces trimmed", "  KEY  =  value  ", "KEY", "value", true},
		{"empty line", "", "", "", false},
		{"whitespace line", "   \t  ", "", "", false},
		{"comment line", "# KEY=value", "", "", false},
		{"no equals sign", "KEYvalue", "", "", false},
		{"empty key", "=value", "", "", false},
		{"double quotes stripped", `KEY="value"`, "KEY", "value", true},
		{"single quotes stripped", "KEY='value'", "KEY", "value", true},
		{"value keeps inner equals", "KEY=a=b=c", "KEY", "a=b=c", true},
		{"empty value", "KEY=", "KEY", "", true},
		{"quoted empty value", `KEY=""`, "KEY", "", true},
		{"inner apostrophe kept", "KEY=don't", "KEY", "don't", true},
		// The dialect deliberately has no inline comments: everything after
		// '=' belongs to the value.
		{"no inline comments", "KEY=value # note", "KEY", "value # note", true},
		{"no export support", "export KEY=value", "export KEY", "value", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, value, ok := parseEnvLine(tt.line)
			if key != tt.wantKey || value != tt.wantValue || ok != tt.wantOK {
				t.Errorf("parseEnvLine(%q) = (%q, %q, %v), want (%q, %q, %v)",
					tt.line, key, value, ok, tt.wantKey, tt.wantValue, tt.wantOK)
			}
		})
	}
}

func TestParseEnvReportsScannerError(t *testing.T) {
	// bufio.Scanner refuses single lines above its buffer limit; parseEnv
	// must surface that instead of silently truncating the file.
	huge := "KEY=" + strings.Repeat("x", 1024*1024)
	err := parseEnv(map[string]any{}, strings.NewReader(huge))
	if err == nil {
		t.Fatal("expected an error for a line exceeding the scanner limit")
	}
}

func FuzzParseEnvLine(f *testing.F) {
	f.Add("KEY=value")
	f.Add("# comment")
	f.Add("  KEY  =  'v'  ")
	f.Add(`A="b=c" # d`)
	f.Add("=")
	f.Add("\"'\"'=\"'")

	f.Fuzz(func(t *testing.T, line string) {
		key, value, ok := parseEnvLine(line)
		if !ok {
			if key != "" || value != "" {
				t.Errorf("not-ok result must be empty, got (%q, %q)", key, value)
			}
			return
		}
		if key == "" {
			t.Error("ok result must have a non-empty key")
		}
		if key != strings.TrimSpace(key) {
			t.Errorf("key %q not trimmed", key)
		}
		if strings.Contains(key, "=") {
			t.Errorf("key %q contains '='", key)
		}
		if len(value) > 0 && (value[0] == '"' || value[0] == '\'' ||
			value[len(value)-1] == '"' || value[len(value)-1] == '\'') {
			t.Errorf("value %q keeps surrounding quotes", value)
		}
	})
}
