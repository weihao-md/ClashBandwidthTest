package main

import "testing"

func TestWorkflowSamplingOnlyWhenSubsetNeeded(t *testing.T) {
	for _, c := range []struct {
		scope, per int
		prescan    bool
	}{
		{scopeSmart, 2, true}, {scopeRegions, 2, true}, {scopeRegions, 0, false},
		{scopeSmart, 0, false}, {scopeFilter, 3, false}, {scopeAll, 2, false},
	} {
		if got := needsLatencyPrescan(c.scope, c.per); got != c.prescan {
			t.Fatalf("scope=%d per=%d got=%v want=%v", c.scope, c.per, got, c.prescan)
		}
	}
}
