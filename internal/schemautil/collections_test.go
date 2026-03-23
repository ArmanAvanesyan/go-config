package schemautil

import (
	"reflect"
	"testing"
)

func TestIsStringMap(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   reflect.Type
		want bool
	}{
		{name: "nil", in: nil, want: false},
		{name: "string map", in: reflect.TypeOf(map[string]int{}), want: true},
		{name: "ptr string map", in: reflect.TypeOf(&map[string]int{}), want: true},
		{name: "non-string key map", in: reflect.TypeOf(map[int]string{}), want: false},
		{name: "slice", in: reflect.TypeOf([]int{}), want: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := IsStringMap(tc.in); got != tc.want {
				t.Fatalf("IsStringMap(%v)=%v want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestIsSliceOrArray(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   reflect.Type
		want bool
	}{
		{name: "nil", in: nil, want: false},
		{name: "slice", in: reflect.TypeOf([]int{}), want: true},
		{name: "array", in: reflect.TypeOf([2]int{}), want: true},
		{name: "ptr slice", in: reflect.TypeOf(&[]int{}), want: true},
		{name: "map", in: reflect.TypeOf(map[string]int{}), want: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := IsSliceOrArray(tc.in); got != tc.want {
				t.Fatalf("IsSliceOrArray(%v)=%v want %v", tc.in, got, tc.want)
			}
		})
	}
}
