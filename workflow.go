package main

const (
	scopeSmart = iota
	scopeRegions
	scopeFilter
	scopeAll
)

// needsLatencyPrescan only ranks latency when it will exclude some candidates.
func needsLatencyPrescan(scope, perRegion int) bool {
	return perRegion > 0 && (scope == scopeSmart || scope == scopeRegions)
}
