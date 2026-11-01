package main

import (
	"errors"
	"fmt"
)

// A successful result pins one origin per batch to preserve comparable speeds.
// Failed origin probes before the first success do not create a leaderboard row.
type sourcePolicy struct {
	Requested string
	Selected  string
	Pinned    bool
	CustomURL string
}

func newSourcePolicy(requested string) *sourcePolicy { return newSourcePolicyWithCustom(requested, "") }
func newSourcePolicyWithCustom(requested, customURL string) *sourcePolicy {
	switch requested {
	case "google", "google_msi", "hetzner", "ovh", "cloudflare", "custom", "auto":
	default:
		requested = "auto"
	}
	selected := requested
	if requested == "auto" {
		selected = "google"
	}
	return &sourcePolicy{Requested: requested, Selected: selected, CustomURL: customURL}
}

func (p *sourcePolicy) Try(run func(downloadSource) (speedResult, error)) (speedResult, error) {
	options := []string{p.Selected}
	if p.Requested == "auto" && !p.Pinned {
		// Cloudflare and OVH deliberately excluded from the automatic path:
		// both have produced repeated HTTP 429 during batch node testing.
		options = []string{"google", "google_msi", "hetzner"}
	}
	var errs []error
	var last speedResult
	for _, key := range options {
		var src downloadSource
		if key == "custom" {
			var err error
			src, err = customDownloadSource(p.CustomURL)
			if err != nil {
				return last, err
			}
		} else {
			src = sourceByName(key)
		}
		result, err := run(src)
		last = result
		if err == nil {
			if !p.Pinned {
				p.Selected = key
				p.Pinned = true
			}
			return result, nil
		}
		if result.RouteFound && !result.RouteOK {
			return result, err
		}
		errs = append(errs, fmt.Errorf("%s：%w", src.Name, err))
	}
	return last, errors.Join(errs...)
}
