package cfgo

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// loadEnvFiles loads the configured env files into m, later files overriding
// earlier ones. With default options the order is .env (base) <
// .{APP_ENV}.env (environment-specific) < .local.env (personal overrides).
// Missing files — and a missing directory for WithDir — are skipped; read
// errors are reported.
func loadEnvFiles(m map[string]any, o options) error {
	files := o.envFiles
	if len(files) == 0 {
		env := os.Getenv(o.envVar)
		if env == "" {
			env = defaultEnv
		}
		files = []string{".env", fmt.Sprintf(".%s.env", env), ".local.env"}
	}

	open := func(name string) (io.ReadCloser, error) { return os.Open(name) }
	if o.dir != "" {
		root, err := os.OpenRoot(o.dir)
		if err != nil {
			return nil // missing directory behaves like missing files
		}
		defer func() { _ = root.Close() }()
		open = func(name string) (io.ReadCloser, error) { return root.Open(name) }
	}

	for _, name := range files {
		if err := loadEnvFile(m, open, name); err != nil {
			return fmt.Errorf("cfgo: env file %q: %w", name, err)
		}
	}
	return nil
}

// loadEnvFile loads a single env file into m; a file that cannot be opened
// is skipped.
func loadEnvFile(m map[string]any, open func(string) (io.ReadCloser, error), filename string) error {
	file, err := open(filename)
	if err != nil {
		return nil // file doesn't exist, skip
	}
	defer func() { _ = file.Close() }()

	return parseEnv(m, file)
}

// parseEnv reads KEY=value lines from r into m.
func parseEnv(m map[string]any, r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if key, value, ok := parseEnvLine(scanner.Text()); ok {
			m[key] = value
		}
	}
	return scanner.Err()
}

// parseEnvLine parses one line of the env-file dialect:
//
//   - blank lines and lines starting with # are skipped
//   - everything before the first '=' is the key, trimmed of whitespace
//   - the value is trimmed of whitespace, then of any leading/trailing
//     single or double quotes
//   - there is no support for `export` prefixes, inline comments, escape
//     sequences, or multiline values
func parseEnvLine(line string) (key, value string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false
	}

	k, v, found := strings.Cut(line, "=")
	if !found {
		return "", "", false
	}

	key = strings.TrimSpace(k)
	if key == "" {
		return "", "", false
	}

	value = strings.Trim(strings.TrimSpace(v), `"'`)
	return key, value, true
}

// loadSystemEnv loads system environment variables into m.
func loadSystemEnv(m map[string]any) {
	for _, env := range os.Environ() {
		if k, v, ok := strings.Cut(env, "="); ok && k != "" {
			m[k] = v
		}
	}
}
