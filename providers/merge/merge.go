package merge

// Strategy represents a configuration tree merge strategy.
//
// Implementations should treat maps as trees and decide how to combine
// existing and new values for each key.
type Strategy interface {
	Merge(dst, src map[string]any) (map[string]any, error)
}
