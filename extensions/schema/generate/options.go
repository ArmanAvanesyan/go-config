package generate

// Options contains resolved generator options.
type Options struct {
	Title       string
	Description string
	SchemaURL   string
	UseRefs     bool
}

// Option configures schema generation.
type Option func(*Options)

// WithTitle sets the root schema title.
func WithTitle(title string) Option {
	return func(o *Options) {
		o.Title = title
	}
}

// WithDescription sets the root schema description.
func WithDescription(desc string) Option {
	return func(o *Options) {
		o.Description = desc
	}
}

// WithSchemaURL sets the $schema meta field (e.g. draft 2020-12).
func WithSchemaURL(url string) Option {
	return func(o *Options) {
		o.SchemaURL = url
	}
}

// WithRefs enables use of $defs and $ref for nested structs instead of
// inlining all object schemas.
func WithRefs(use bool) Option {
	return func(o *Options) {
		o.UseRefs = use
	}
}

// ResolveOptions applies options and returns resolved values with defaults.
func ResolveOptions(opts ...Option) Options {
	o := Options{
		SchemaURL: "https://json-schema.org/draft/2020-12/schema",
	}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
