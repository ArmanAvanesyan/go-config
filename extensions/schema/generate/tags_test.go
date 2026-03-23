package generate

import "testing"

func TestParseJSONTag_Table(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want fieldTag
	}{
		{"-", fieldTag{Skip: true}},
		{"", fieldTag{}},
		{"name", fieldTag{Name: "name"}},
		{"name,omitempty", fieldTag{Name: "name", OmitEmpty: true}},
		{" name , omitempty , unknown ", fieldTag{Name: "name", OmitEmpty: true}},
		{",omitempty", fieldTag{OmitEmpty: true}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			got := parseJSONTag(tc.in)
			if got != tc.want {
				t.Fatalf("parseJSONTag(%q)=%+v want %+v", tc.in, got, tc.want)
			}
		})
	}
}
