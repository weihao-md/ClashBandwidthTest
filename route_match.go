package main

import (
	"encoding/json"
	"net"
	"strconv"
	"strings"
)

// Mihomo's /connections endpoint may serialize sourcePort as a JSON number
// or a JSON string; preserve both representations for reliable correlation.
type connectionEntry struct {
	ID       string   `json:"id"`
	Download int64    `json:"download"`
	Chains   []string `json:"chains"`
	Rule     string   `json:"rule"`
	Metadata struct {
		Host          string          `json:"host"`
		DestinationIP string          `json:"destinationIP"`
		SourceIP      string          `json:"sourceIP"`
		SourcePort    json.RawMessage `json:"sourcePort"`
	} `json:"metadata"`
}

type connectionSnapshot struct {
	Connections []connectionEntry `json:"connections"`
}

func parseSourcePort(raw json.RawMessage) (int, bool) {
	v := strings.TrimSpace(string(raw))
	if len(v) == 0 {
		return 0, false
	}
	if v[0] == '"' {
		var decoded string
		if json.Unmarshal(raw, &decoded) != nil {
			return 0, false
		}
		v = decoded
	}
	p, err := strconv.Atoi(v)
	return p, err == nil && p > 0 && p <= 65535
}

// matchOwnProxyConnection selects ONLY the connection initiated by this exact
// local proxy socket. Host-only / first-new-ID matching is unsafe when other
// benchmark traffic exists concurrently.
func matchOwnProxyConnection(snap connectionSnapshot, host string, localPort int) (connectionEntry, bool) {
	for _, c := range snap.Connections {
		port, ok := parseSourcePort(c.Metadata.SourcePort)
		if !ok || port != localPort {
			continue
		}
		if c.Metadata.SourceIP != "" {
			ip := net.ParseIP(strings.Trim(c.Metadata.SourceIP, "[]"))
			if ip == nil || !ip.IsLoopback() {
				continue
			}
		}
		if c.Metadata.Host != "" && !strings.EqualFold(strings.TrimSuffix(c.Metadata.Host, "."), host) {
			continue
		}
		return c, true
	}
	var empty connectionEntry
	return empty, false
}
