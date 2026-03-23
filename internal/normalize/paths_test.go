package normalize

import "testing"

func TestPathAndEnvToPath(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"  ", ""},
		{" Server.Port ", "server.port"},
		{"A.__B", "a.__b"},
	}
	for _, tc := range cases {
		if got := Path(tc.in); got != tc.want {
			t.Fatalf("Path(%q)=%q want=%q", tc.in, got, tc.want)
		}
	}

	if got := EnvToPath(" APP__SERVER_PORT "); got != "app.server.port" {
		t.Fatalf("EnvToPath mismatch: %q", got)
	}
}
