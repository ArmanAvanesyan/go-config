package mapstructure

import internaldecode "github.com/ArmanAvanesyan/go-config/internal/decode"

// Decoder decodes a generic config map into a struct using mapstructure.
type Decoder struct{}

// New returns a new mapstructure-based Decoder.
func New() *Decoder {
	return &Decoder{}
}

// Decode decodes the generic config tree into out without JSON re-marshaling.
func (d *Decoder) Decode(input map[string]any, out any) error {
	return internaldecode.Decode(input, out, internaldecode.Options{
		WeaklyTypedInput: true,
		ErrorUnused:      false,
	})
}
