package config

import "errors"

var (
	// ErrNilTarget is returned when Load is called with a nil target.
	ErrNilTarget = errors.New("config: target must not be nil")
	// ErrNoSources is returned when Load is called with no sources added.
	ErrNoSources = errors.New("config: no sources configured")
	// ErrUnknownFormat is returned when the document format cannot be determined.
	ErrUnknownFormat = errors.New("config: unknown document format")
	// ErrParserRequired is returned when a raw document has no parser configured.
	ErrParserRequired = errors.New("config: parser required for raw documents")
	// ErrDecoderRequired is returned when no decoder is configured.
	ErrDecoderRequired = errors.New("config: decoder required")
	// ErrInvalidDocument is returned when the document is invalid or empty.
	ErrInvalidDocument = errors.New("config: invalid document")
	// ErrValidationFailed is returned when validation fails.
	ErrValidationFailed = errors.New("config: validation failed")
	// ErrResolutionFailed is returned when placeholder resolution fails.
	ErrResolutionFailed = errors.New("config: resolution failed")
	// ErrDecodeFailed is returned when decoding the tree into the target fails.
	ErrDecodeFailed = errors.New("config: decode failed")
	// ErrSourceReadFailed is returned when a source fails to read.
	ErrSourceReadFailed = errors.New("config: source read failed")
	// ErrParseFailed is returned when parsing a document fails.
	ErrParseFailed = errors.New("config: parse failed")
	// ErrMergeFailed is returned when merging multiple sources fails.
	ErrMergeFailed = errors.New("config: merge failed")
)
