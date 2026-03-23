package config

// Document represents raw source material before parsing.
type Document struct {
	Name     string
	Format   string
	Raw      []byte
	Metadata Metadata
}

// TreeDocument represents already-structured source data.
type TreeDocument struct {
	Name     string
	Tree     map[string]any
	Metadata Metadata
}
