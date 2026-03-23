package generate

import "strings"

// fieldTag describes how a struct field's `json` tag should be interpreted.
type fieldTag struct {
	Name      string
	OmitEmpty bool
	Skip      bool
}

// parseJSONTag parses a struct field's `json` tag into fieldTag.
func parseJSONTag(tag string) fieldTag {
	if tag == "-" {
		return fieldTag{Skip: true}
	}
	if tag == "" {
		return fieldTag{}
	}

	parts := strings.Split(tag, ",")
	ft := fieldTag{}

	if name := strings.TrimSpace(parts[0]); name != "" {
		ft.Name = name
	}

	for _, opt := range parts[1:] {
		switch strings.TrimSpace(opt) {
		case "omitempty":
			ft.OmitEmpty = true
		}
	}

	return ft
}
