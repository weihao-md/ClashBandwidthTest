package main

import (
	"encoding/json"
	"testing"
)

func TestParseSourcePort(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		valid    bool
	}{
		{`"60321"`, 60321, true}, {`60321`, 60321, true},
		{`0`, 0, false}, {`65536`, 65536, false}, {`null`, 0, false},
		{`"abc"`, 0, false},
	}
	for _, tt := range tests {
		got, ok := parseSourcePort(json.RawMessage(tt.input))
		if got != tt.expected || ok != tt.valid {
			t.Errorf("parseSourcePort(%s)=%d,%t; want %d,%t", tt.input, got, ok, tt.expected, tt.valid)
		}
	}
}

func TestMatchOwnProxyConnectionNeverUsesFirstHostMatch(t *testing.T) {
	raw := `{"connections":[
      {"id":"wrong","chains":["日本 BA D"],"metadata":{"host":"speed.cloudflare.com","sourceIP":"127.0.0.1","sourcePort":"50001"}},
      {"id":"ours","chains":["日本 SK A"],"metadata":{"host":"speed.cloudflare.com","sourceIP":"127.0.0.1","sourcePort":50002}},
      {"id":"otherhost","metadata":{"host":"example.com","sourceIP":"127.0.0.1","sourcePort":50003}}
    ]}`
	var snap connectionSnapshot
	if err := json.Unmarshal([]byte(raw), &snap); err != nil {
		t.Fatal(err)
	}
	got, ok := matchOwnProxyConnection(snap, "speed.cloudflare.com", 50002)
	if !ok || got.ID != "ours" {
		t.Fatalf("expected ours, got %+v %t", got, ok)
	}
	if _, ok := matchOwnProxyConnection(snap, "speed.cloudflare.com", 50003); ok {
		t.Fatal("different hostname must not be attributed to benchmark")
	}
	if _, ok := matchOwnProxyConnection(snap, "speed.cloudflare.com", 50004); ok {
		t.Fatal("unrelated source port must not be attributed to benchmark")
	}
}

func TestRejectNonLocalSource(t *testing.T) {
	raw := `{"connections":[{"id":"remote","metadata":{"host":"speed.cloudflare.com","sourceIP":"8.8.8.8","sourcePort":"50002"}}]}`
	var snap connectionSnapshot
	if err := json.Unmarshal([]byte(raw), &snap); err != nil {
		t.Fatal(err)
	}
	if _, ok := matchOwnProxyConnection(snap, "speed.cloudflare.com", 50002); ok {
		t.Fatal("must not match a remote proxy connection")
	}
}
