package testutil

import (
	"errors"
	"reflect"
	"testing"
)

// RequireNoError marks the test as failed and stops execution if err is non-nil.
func RequireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// RequireEqual fails with a diff-friendly message if want != got.
func RequireEqual(t *testing.T, want, got any) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("mismatch:\n  want: %#v\n   got: %#v", want, got)
	}
}

// RequireError fails if err is nil.
func RequireError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// RequireErrorIs fails if !errors.Is(err, target).
func RequireErrorIs(t *testing.T, err, target error) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatalf("expected errors.Is(%v, %v) to be true", err, target)
	}
}

// MustSetEnv sets key=value for the duration of the test via t.Setenv.
func MustSetEnv(t *testing.T, key, value string) {
	t.Helper()
	t.Setenv(key, value)
}
