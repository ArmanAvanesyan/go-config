package generate

import "testing"

func TestResolveOptions_Table(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		opts []Option
		want Options
	}{
		{
			name: "defaults",
			want: Options{SchemaURL: "https://json-schema.org/draft/2020-12/schema"},
		},
		{
			name: "all options applied",
			opts: []Option{
				WithTitle("title"),
				WithDescription("desc"),
				WithSchemaURL("schema"),
				WithRefs(true),
			},
			want: Options{
				Title:       "title",
				Description: "desc",
				SchemaURL:   "schema",
				UseRefs:     true,
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := ResolveOptions(tc.opts...)
			if got != tc.want {
				t.Fatalf("ResolveOptions()=%+v want %+v", got, tc.want)
			}
		})
	}
}
