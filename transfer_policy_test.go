package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestWarmupAssessment(t *testing.T) {
	cases := []struct {
		name             string
		want, got        int64
		err              error
		complete, usable bool
	}{
		{"complete", 1000000, 1000000, nil, true, true},
		{"interrupted-but-usable", 1000000, 482531, io.ErrUnexpectedEOF, false, true},
		{"too-few", 1000000, 2000, io.ErrUnexpectedEOF, false, false},
		{"early-eof", 1000000, 500000, io.EOF, false, true},
		{"exact-EOF", 1000000, 1000000, io.EOF, true, true},
		{"exact-unexpected-EOF", 1000000, 1000000, io.ErrUnexpectedEOF, true, true},
		{"other-error", 1000000, 200000, errors.New("connection reset"), false, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			complete, usable, reason := warmupAssessment(c.want, c.got, c.err)
			if complete != c.complete || usable != c.usable {
				t.Fatalf("want (%v,%v), got (%v,%v) [reason=%s]", c.complete, c.usable, complete, usable, reason)
			}
			if !complete && !strings.Contains(reason, "预热") {
				t.Fatalf("missing warning: %q", reason)
			}
		})
	}
}

func TestPartialBenchmark(t *testing.T) {
	for _, c := range []struct {
		got, want      int
		early, partial bool
	}{{2, 2, false, false}, {1, 2, false, true}, {2, 2, true, true}, {0, 2, false, true}} {
		if x := partialBenchmark(c.got, c.want, c.early); x != c.partial {
			t.Fatalf("%+v: got %v", c, x)
		}
	}
}

func TestReadAtMostIncompleteEOF(t *testing.T) {
	got, err := readAtMost(context.Background(), bytes.NewReader(make([]byte, 482531)), 1000000)
	if got != 482531 || !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("got %d, %v", got, err)
	}
}
func TestReadAtMostComplete(t *testing.T) {
	got, err := readAtMost(context.Background(), bytes.NewReader(make([]byte, 1000000)), 1000000)
	if got != 1000000 || err != nil {
		t.Fatalf("got %d, %v", got, err)
	}
}
