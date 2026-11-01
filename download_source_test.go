package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestCustomDownloadSource(t *testing.T) {
	for _, bad := range []string{"http://example.org/file", "https://example.com", "https://name:pwd@example.org/test", "https://example.org/file#frag", "bad"} {
		if _, err := customDownloadSource(bad); err == nil {
			t.Fatalf("accepted %q", bad)
		}
	}
	src, err := customDownloadSource("https://example.org/file.bin?size=123")
	if err != nil || src.URL(100) != "https://example.org/file.bin?size=123" || src.Host != "example.org" {
		t.Fatalf("bad custom %+v %v", src, err)
	}
}
func TestGoogleFirstAndCustomSourcePinned(t *testing.T) {
	p := newSourcePolicy("auto")
	keys := []string{}
	_, err := p.Try(func(s downloadSource) (speedResult, error) {
		keys = append(keys, s.Key)
		return speedResult{Source: s.Name, Average: 2}, nil
	})
	if err != nil || !p.Pinned || p.Selected != "google" || len(keys) != 1 {
		t.Fatalf("auto key %v %v", keys, err)
	}
	c := newSourcePolicyWithCustom("custom", "https://example.com/file.bin")
	x, n := c.Try(func(s downloadSource) (speedResult, error) {
		if !strings.HasPrefix(s.URL(1), "https://example.com/") {
			return speedResult{}, fmt.Errorf("wrong url")
		}
		return speedResult{Source: s.Name}, nil
	})
	if n != nil || x.Source != "自定义文件" || !c.Pinned {
		t.Fatalf("custom %+v %v", x, n)
	}
}

func TestCustomHostWithoutPort(t *testing.T) {
	src, err := customDownloadSource("https://example.org:8443/package.iso")
	if err != nil || src.Host != "example.org" || !strings.Contains(src.URL(1), ":8443/") {
		t.Fatalf("custom host %+v %v", src, err)
	}
}
