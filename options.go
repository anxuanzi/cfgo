package cfgo

// defaultEnv is the environment assumed when the selector variable is unset.
const defaultEnv = "dev"

// options holds the loading configuration for a Config instance.
type options struct {
	dir              string
	envFiles         []string
	includeSystemEnv bool
	envVar           string
}

func defaultOptions() options {
	return options{
		includeSystemEnv: true,
		envVar:           "APP_ENV",
	}
}

// Option customizes how New loads configuration.
type Option func(*options)

// WithDir loads env files from dir instead of the current working directory.
// The directory is opened with os.Root, so file lookups cannot escape it via
// path traversal or symlinks. A missing directory is treated like missing
// files.
func WithDir(dir string) Option {
	return func(o *options) { o.dir = dir }
}

// WithEnvFiles replaces the default file set (.env, .{APP_ENV}.env,
// .local.env) with the given files, loaded in order; later files override
// earlier ones. Names are used verbatim — no environment-name substitution
// is applied.
func WithEnvFiles(files ...string) Option {
	return func(o *options) { o.envFiles = append([]string(nil), files...) }
}

// WithoutSystemEnv excludes process environment variables from the
// configuration. This also keeps secrets held in the environment out of All
// and Each.
func WithoutSystemEnv() Option {
	return func(o *options) { o.includeSystemEnv = false }
}

// WithEnvVar changes the environment variable that selects the
// environment-specific file (default "APP_ENV").
func WithEnvVar(name string) Option {
	return func(o *options) {
		if name != "" {
			o.envVar = name
		}
	}
}
