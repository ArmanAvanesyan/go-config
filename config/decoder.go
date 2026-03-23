package config

// Decoder decodes a merged config tree into a typed struct.
type Decoder interface {
	Decode(input map[string]any, out any) error
}
