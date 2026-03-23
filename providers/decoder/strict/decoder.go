package strict

import internaldecode "github.com/ArmanAvanesyan/go-config/internal/decode"

// Decoder decodes a config map into a struct using strict JSON (unknown fields cause an error).
type Decoder struct{}

// New returns a new strict Decoder.
func New() *Decoder {
	return &Decoder{}
}

// Decode decodes the generic config tree strictly; unknown fields fail.
func (d *Decoder) Decode(input map[string]any, out any) error {
	return internaldecode.Decode(input, out, internaldecode.Options{
		WeaklyTypedInput: false,
		ErrorUnused:      true,
	})
}
