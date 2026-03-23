package normalize

import "strings"

// Path canonicalizes a dot-separated path by normalizing each segment.
func Path(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	parts := strings.Split(path, ".")
	for i := range parts {
		parts[i] = Key(parts[i])
	}
	return strings.Join(parts, ".")
}

// EnvToPath converts env-style keys into dot paths.
// Example: APP_SERVER__PORT -> app.server.port (after prefix stripping by caller).
func EnvToPath(key string) string {
	normalized := strings.ReplaceAll(strings.TrimSpace(key), "__", ".")
	normalized = strings.ReplaceAll(normalized, "_", ".")
	return Path(normalized)
}
