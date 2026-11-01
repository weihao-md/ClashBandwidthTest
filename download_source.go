package main

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type downloadSource struct {
	Key           string
	Name          string
	Host          string
	Path          string
	FullURL       string
	SupportsRange bool
}

// Static large file sources are deliberately preferred to specialized speedtest
// endpoints which frequently limit scripted requests. The Google Chrome URL
// matches the default used by the independently maintained clash-speedtest
// project; no code from that GPL-3.0 project is included here.
func sourceByName(name string) downloadSource {
	switch name {
	case "google":
		return downloadSource{Key: "google", Name: "Google Chrome 文件", Host: "dl.google.com", Path: "/chrome/mac/universal/stable/GGRO/googlechrome.dmg", SupportsRange: true}
	case "google_msi":
		return downloadSource{Key: "google_msi", Name: "Google Chrome Windows 文件", Host: "dl.google.com", Path: "/dl/chrome/install/googlechromestandaloneenterprise64.msi", SupportsRange: true}
	case "hetzner":
		return downloadSource{Key: "hetzner", Name: "Hetzner 德国", Host: "fsn1-speed.hetzner.com", Path: "/100MB.bin", SupportsRange: true}
	case "ovh":
		return downloadSource{Key: "ovh", Name: "OVH 新加坡", Host: "sin1-sgp.speedtest.network.ovh.net", Path: "/files/100Mb.dat", SupportsRange: true}
	default:
		return downloadSource{Key: "cloudflare", Name: "Cloudflare", Host: "speed.cloudflare.com", Path: "/__down"}
	}
}

// Custom file URLs must be HTTPS, have a hostname and a real file path, and
// must not embed URL credentials. No remote URL is ever persisted by this app.
func customDownloadSource(raw string) (downloadSource, error) {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil || u == nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Fragment != "" || u.Opaque != "" {
		return downloadSource{}, fmt.Errorf("自定义测速文件必须是 HTTPS URL，不得包含账号密码或 #片段")
	}
	if u.Path == "" || u.Path == "/" {
		return downloadSource{}, fmt.Errorf("请填写指向真实大文件的 URL（含文件路径）")
	}
	return downloadSource{Key: "custom", Name: "自定义文件", Host: u.Hostname(), FullURL: u.String(), SupportsRange: true}, nil
}

func (s downloadSource) URL(bytes int64) string {
	if s.FullURL != "" {
		return s.FullURL
	}
	if s.SupportsRange {
		path := s.Path
		if path == "" {
			path = "/files/100Mb.dat"
		} // existing local HTTP range regression harness
		return "https://" + s.Host + path
	}
	path := s.Path
	if path == "" {
		path = "/__down"
	}
	return "https://" + s.Host + path + "?" + url.Values{"bytes": {strconv.FormatInt(bytes, 10)}}.Encode()
}
