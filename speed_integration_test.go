package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func localConnectProxy(t *testing.T) (*httptest.Server, int) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodConnect {
			http.Error(w, "CONNECT only", 405)
			return
		}
		target, err := net.DialTimeout("tcp", r.Host, 2*time.Second)
		if err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
		h, ok := w.(http.Hijacker)
		if !ok {
			target.Close()
			http.Error(w, "no hijack", 500)
			return
		}
		incoming, buf, err := h.Hijack()
		if err != nil {
			target.Close()
			return
		}
		_, _ = buf.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n")
		if err := buf.Flush(); err != nil {
			incoming.Close()
			target.Close()
			return
		}
		go func() { defer incoming.Close(); defer target.Close(); _, _ = io.Copy(target, buf) }()
		go func() { defer incoming.Close(); defer target.Close(); _, _ = io.Copy(incoming, target) }()
	}))
	u, _ := url.Parse(server.URL)
	port, _ := strconv.Atoi(u.Port())
	return server, port
}
func testClientFactory(t *testing.T, server *httptest.Server) func(int) (*http.Client, *http.Transport, *atomic.Int64) {
	t.Helper()
	ca := x509.NewCertPool()
	ca.AddCert(server.Certificate())
	return func(port int) (*http.Client, *http.Transport, *atomic.Int64) {
		c, tr, p := newSpeedClient(port)
		tr.TLSClientConfig = &tls.Config{RootCAs: ca}
		return c, tr, p
	}
}
func runLocalBenchmark(t *testing.T, srv *httptest.Server, profile testProfile, static bool) (speedResult, error) {
	t.Helper()
	proxy, port := localConnectProxy(t)
	defer proxy.Close()
	u, _ := url.Parse(srv.URL)
	src := downloadSource{Name: "Mock", Host: u.Host, SupportsRange: static}
	f := testClientFactory(t, srv)
	return benchmarkDownload(context.Background(), "MockNode", port, profile, src, "test", f, func(ctx context.Context, host string, localPort int) (bool, bool, string) {
		if localPort <= 0 {
			t.Error("missing local TCP port")
		}
		if host != u.Host {
			t.Errorf("wrong route host %s", host)
		}
		return true, true, "✓ MockNode"
	})
}
func TestHTTPSViaCONNECTProxyFullBenchmark(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		size, _ := strconv.Atoi(r.URL.Query().Get("bytes"))
		if r.URL.Path != "/__down" || size <= 0 {
			http.Error(w, "wrong request", 400)
			return
		}
		w.Header().Set("Content-Length", strconv.Itoa(size))
		w.WriteHeader(200)
		_, _ = io.CopyN(w, strings.NewReader(strings.Repeat("x", size)), int64(size))
	}))
	defer srv.Close()
	p := testProfile{WarmupMB: 1, Streams: 2, PerStreamMB: 1, Duration: 300 * time.Millisecond}
	r, err := runLocalBenchmark(t, srv, p, false)
	if err != nil {
		t.Fatalf("full benchmark failed: %v", err)
	}
	if r.Route != "✓ MockNode" || r.Source != "Mock" || r.Average <= 0 {
		t.Fatalf("bad result %+v", r)
	}
}
func TestStaticRangeDoesNotOvershoot(t *testing.T) {
	var fullBytes atomic.Int64
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/files/100Mb.dat" {
			http.Error(w, "wrong path", 400)
			return
		}
		value := r.Header.Get("Range")
		if !strings.HasPrefix(value, "bytes=0-") {
			http.Error(w, "range required", 400)
			return
		}
		end, err := strconv.Atoi(strings.TrimPrefix(value, "bytes=0-"))
		if err != nil {
			http.Error(w, "bad range", 400)
			return
		}
		n := end + 1
		fullBytes.Add(int64(n))
		w.Header().Set("Content-Length", strconv.Itoa(n))
		w.WriteHeader(206)
		_, _ = io.CopyN(w, strings.NewReader(strings.Repeat("r", n)), int64(n))
	}))
	defer srv.Close()
	r, err := runLocalBenchmark(t, srv, testProfile{WarmupMB: 1, Streams: 2, PerStreamMB: 1, Duration: 250 * time.Millisecond}, true)
	if err != nil {
		t.Fatalf("range benchmark failed: %v", err)
	}
	if r.Average <= 0 {
		t.Fatalf("no output %+v", r)
	}
	if fullBytes.Load() != 2000000 {
		t.Fatalf("server sent %d bytes, expected 2000000", fullBytes.Load())
	}
}
func TestTruncatedWarmupRetryNotClean(t *testing.T) {
	var calls atomic.Int64
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Length", "1000000")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(strings.Repeat("q", 482531)))
	}))
	defer srv.Close()
	r, err := runLocalBenchmark(t, srv, testProfile{WarmupMB: 1, Streams: 2, PerStreamMB: 1, Duration: 300 * time.Millisecond}, false)
	if calls.Load() < 1 {
		t.Errorf("expected a real download request, got %d", calls.Load())
	}
	if err == nil && r.Warning == "" {
		t.Fatalf("truncated transfer was marked clean: %+v", r)
	}
}
func TestTimeoutStopsBenchmark(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { time.Sleep(300 * time.Millisecond); w.WriteHeader(200) }))
	defer srv.Close()
	proxy, port := localConnectProxy(t)
	defer proxy.Close()
	u, _ := url.Parse(srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := benchmarkDownload(ctx, "MockNode", port, testProfile{WarmupMB: 1, Streams: 2, PerStreamMB: 1, Duration: 200 * time.Millisecond}, downloadSource{Name: "Mock", Host: u.Host}, "test", testClientFactory(t, srv), func(context.Context, string, int) (bool, bool, string) { return false, false, "?" })
	if err == nil {
		t.Fatal("expected timeout error")
	}
}
func TestSourcePolicyBatchPinsFallback(t *testing.T) {
	p := newSourcePolicy("auto")
	var calls []string
	run := func(src downloadSource) (speedResult, error) {
		calls = append(calls, src.Name)
		if src.Key == "google" {
			return speedResult{Source: src.Name}, fmt.Errorf("OVH unavailable")
		}
		return speedResult{Source: src.Name, Average: 3.0}, nil
	}
	result, err := p.Try(run)
	if err != nil || result.Source != "Google Chrome Windows 文件" {
		t.Fatalf("fallback: %+v %v", result, err)
	}
	_, err = p.Try(run)
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 3 || calls[0] != "Google Chrome 文件" || calls[1] != "Google Chrome Windows 文件" || calls[2] != "Google Chrome Windows 文件" {
		t.Fatalf("batch mix: %v", calls)
	}
}
func TestSourcePolicyThirdFallbackAndSourceIsExplicit(t *testing.T) {
	p := newSourcePolicy("auto")
	var calls []string
	res, err := p.Try(func(src downloadSource) (speedResult, error) {
		calls = append(calls, src.Key)
		if src.Key != "hetzner" {
			return speedResult{}, fmt.Errorf("unavailable")
		}
		return speedResult{Average: 1, Source: src.Name}, nil
	})
	if err != nil || res.Source != "Hetzner 德国" || p.Selected != "hetzner" || len(calls) != 3 {
		t.Fatalf("bad fallback %+v %v %v", res, err, calls)
	}
}
func TestManualSourceNeverFallsBack(t *testing.T) {
	p := newSourcePolicy("hetzner")
	n := 0
	_, err := p.Try(func(src downloadSource) (speedResult, error) {
		n++
		return speedResult{}, fmt.Errorf("connection down")
	})
	if err == nil || n != 1 {
		t.Fatalf("manual source should not fallback: %d %v", n, err)
	}
}
func TestSourcePolicyRouteMismatchNoFallback(t *testing.T) {
	p := newSourcePolicy("auto")
	n := 0
	_, err := p.Try(func(src downloadSource) (speedResult, error) {
		n++
		return speedResult{RouteFound: true, RouteOK: false}, fmt.Errorf("wrong route")
	})
	if err == nil || n != 1 {
		t.Fatalf("should not fallback on wrong route; calls=%d err=%v", n, err)
	}
}

func TestStaticServerIgnoringRangeStillCapsClientReads(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/files/100Mb.dat" {
			http.Error(w, "wrong path", 404)
			return
		}
		w.Header().Set("Content-Length", strconv.Itoa(10000000))
		w.WriteHeader(200)
		_, _ = io.CopyN(w, strings.NewReader(strings.Repeat("z", 10000000)), 10000000)
	}))
	defer srv.Close()
	out, err := runLocalBenchmark(t, srv, testProfile{WarmupMB: 1, Streams: 2, PerStreamMB: 1, Duration: 250 * time.Millisecond}, true)
	if err != nil {
		t.Fatalf("ignoring range must still be safe: %v", err)
	}
	if out.Average <= 0 {
		t.Fatalf("invalid result %+v", out)
	}
}

func TestOneStreamFailsReportedPartial(t *testing.T) {
	var opened atomic.Int64
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		size, _ := strconv.Atoi(r.URL.Query().Get("bytes"))
		if size == 2000000 && opened.Add(1) >= 2 {
			http.Error(w, "upstream busy", 503)
			return
		}
		w.Header().Set("Content-Length", strconv.Itoa(size))
		w.WriteHeader(200)
		_, _ = io.CopyN(w, strings.NewReader(strings.Repeat("t", size)), int64(size))
	}))
	defer srv.Close()
	out, err := runLocalBenchmark(t, srv, testProfile{WarmupMB: 1, Streams: 2, PerStreamMB: 2, Duration: 250 * time.Millisecond}, false)
	if err != nil {
		t.Fatal(err)
	}
	if out.Average <= 0 || !strings.Contains(out.Warning, "1/2") {
		t.Fatalf("failure should be labeled partial %+v", out)
	}
}

func TestConfirmedWrongRouteFailsWithoutSpeed(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		size, _ := strconv.Atoi(r.URL.Query().Get("bytes"))
		w.WriteHeader(200)
		_, _ = io.CopyN(w, strings.NewReader(strings.Repeat("y", size)), int64(size))
	}))
	defer srv.Close()
	proxy, port := localConnectProxy(t)
	defer proxy.Close()
	u, _ := url.Parse(srv.URL)
	out, err := benchmarkDownload(context.Background(), "NodeA", port, testProfile{WarmupMB: 1, Streams: 2, PerStreamMB: 1, Duration: 200 * time.Millisecond}, downloadSource{Name: "Mock", Host: u.Host}, "test", testClientFactory(t, srv), func(context.Context, string, int) (bool, bool, string) { return true, false, "⚠ NodeB" })
	if err == nil || !out.RouteFound || out.RouteOK || out.Average > 0 {
		t.Fatalf("must not score wrong route: %+v %v", out, err)
	}
}

// Regression for the reported 0.1MB/s links: preflight must not wait for a full
// 1MB transfer before commencing the 2-second benchmark.
func TestSlowNodeNoFullMegabytePreflight(t *testing.T) {
	var calls atomic.Int64
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Length", strconv.Itoa(6000000))
		w.WriteHeader(http.StatusOK)
		flush, _ := w.(http.Flusher)
		buf := []byte(strings.Repeat("s", 10240))
		for i := 0; i < 250; i++ {
			if _, err := w.Write(buf); err != nil {
				return
			}
			if flush != nil {
				flush.Flush()
			}
			select {
			case <-r.Context().Done():
				return
			case <-time.After(100 * time.Millisecond):
			}
		}
	}))
	defer srv.Close()
	start := time.Now()
	out, err := runLocalBenchmark(t, srv, testProfile{WarmupMB: 1, Streams: 2, PerStreamMB: 6, Duration: 2 * time.Second}, false)
	if err != nil {
		t.Fatalf("slow node should be measurable: %v", err)
	}
	if out.Average < 0.04 || out.Average > 0.4 {
		t.Fatalf("unexpected average %+v", out)
	}
	if time.Since(start) > 12*time.Second {
		t.Fatalf("preflight took too long: %v", time.Since(start))
	}
	if calls.Load() != 2 {
		t.Fatalf("expected only 2 measured streams without extra probe, got %d", calls.Load())
	}
}
func TestDocumentedSourceURLs(t *testing.T) {
	cases := []struct{ key, want string }{
		{"google", "https://dl.google.com/chrome/mac/universal/stable/GGRO/googlechrome.dmg"},
		{"google_msi", "https://dl.google.com/dl/chrome/install/googlechromestandaloneenterprise64.msi"},
		{"ovh", "https://sin1-sgp.speedtest.network.ovh.net/files/100Mb.dat"},
		{"hetzner", "https://fsn1-speed.hetzner.com/100MB.bin"},
		{"cloudflare", "https://speed.cloudflare.com/__down?bytes=16384"},
	}
	for _, tc := range cases {
		if got := sourceByName(tc.key).URL(16384); got != tc.want {
			t.Fatalf("%s URL: %s", tc.key, got)
		}
	}
}

// A source can reject the second concurrent stream after accepting the first.
// This must NEVER become a comparable partial success.
func Test429InSecondStreamIsNotPartial(t *testing.T) {
	var requests atomic.Int64
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requests.Add(1) == 2 {
			http.Error(w, "Too Many Requests", 429)
			return
		}
		size, _ := strconv.Atoi(r.URL.Query().Get("bytes"))
		w.Header().Set("Content-Length", strconv.Itoa(size))
		w.WriteHeader(200)
		_, _ = io.CopyN(w, strings.NewReader(strings.Repeat("z", size)), int64(size))
	}))
	defer srv.Close()
	r, err := runLocalBenchmark(t, srv, testProfile{Streams: 2, PerStreamMB: 1, Duration: 250 * time.Millisecond}, false)
	if !isRateLimited(err) || r.Average > 0 || requests.Load() != 2 {
		t.Fatalf("must fail on 429 without retry or partial score; requests=%d result=%+v err=%v", requests.Load(), r, err)
	}
}
func Test429FirstStreamNoRetry(t *testing.T) {
	var requests atomic.Int64
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		http.Error(w, "rate limit", 429)
	}))
	defer srv.Close()
	r, err := runLocalBenchmark(t, srv, testProfile{Streams: 1, PerStreamMB: 1, Duration: 250 * time.Millisecond}, false)
	if !isRateLimited(err) || r.Average > 0 || requests.Load() != 1 {
		t.Fatalf("429 must not retry requests=%d result=%+v err=%v", requests.Load(), r, err)
	}
}
func TestOneStreamNoExtraPreflight(t *testing.T) {
	var requests atomic.Int64
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		size, _ := strconv.Atoi(r.URL.Query().Get("bytes"))
		w.Header().Set("Content-Length", strconv.Itoa(size))
		w.WriteHeader(200)
		_, _ = io.CopyN(w, strings.NewReader(strings.Repeat("a", size)), int64(size))
	}))
	defer srv.Close()
	r, err := runLocalBenchmark(t, srv, testProfile{Streams: 1, PerStreamMB: 1, Duration: 250 * time.Millisecond}, false)
	if err != nil || requests.Load() != 1 || r.Average <= 0 || !r.RouteOK {
		t.Fatalf("expected one full verified stream requests=%d result=%+v err=%v", requests.Load(), r, err)
	}
}

// Regression: the first stream's body must be canceled when the benchmark
// duration is over, otherwise blocked reads can hang the GUI indefinitely.
func TestBlockedFirstStreamIsCanceledOnTimer(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer server.Close()
	start := time.Now()
	_, err := runLocalBenchmark(t, server, testProfile{Streams: 1, PerStreamMB: 1, Duration: 180 * time.Millisecond}, false)
	if err == nil {
		t.Fatal("zero-byte blocked stream must fail")
	}
	if time.Since(start) > 2*time.Second {
		t.Fatalf("benchmark body read was not canceled: %v", time.Since(start))
	}
}

func TestPinnedRateLimitDoesNotFallbackOrReturnSpeed(t *testing.T) {
	p := newSourcePolicy("auto")
	got, n := 0, 0
	_, err := p.Try(func(s downloadSource) (speedResult, error) { n++; return speedResult{Source: s.Name, Average: 1}, nil })
	if err != nil || !p.Pinned {
		t.Fatal("expected source pin", err)
	}
	_, err = p.Try(func(s downloadSource) (speedResult, error) {
		got++
		return speedResult{Source: s.Name}, &SourceHTTPError{Source: s.Name, Code: 429, Stage: "GET"}
	})
	if !isRateLimited(err) || got != 1 || n != 1 {
		t.Fatalf("pinned 429 must stop with one request: %d %d %v", got, n, err)
	}
}

// A server returning HTTP 200 with an HTML error page is not a successful
// static-file download (common with CDN block pages / captive portals).
func TestHTML200IsRejectedRatherThanMeasured(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "<!doctype html><html><body>blocked</body></html>")
	}))
	defer srv.Close()
	result, err := runLocalBenchmark(t, srv, testProfile{Streams: 1, PerStreamMB: 1, Duration: 200 * time.Millisecond}, false)
	if err == nil || !strings.Contains(err.Error(), "HTML") || result.Average > 0 {
		t.Fatalf("HTML error measured: %+v %v", result, err)
	}
}

// Default auto source should skip providers known to produce rate limits
// in bulk, and a successful static source is pinned for the batch.
func TestAutoAvoids429ProneSourcesAndLocksSource(t *testing.T) {
	p := newSourcePolicy("auto")
	var attempted []string
	_, err := p.Try(func(s downloadSource) (speedResult, error) {
		attempted = append(attempted, s.Key)
		if s.Key == "google" {
			return speedResult{}, fmt.Errorf("first source unreachable")
		}
		return speedResult{Source: s.Name, Average: 1}, nil
	})
	if err != nil || !p.Pinned || p.Selected != "google_msi" {
		t.Fatalf("pin failed %v %+v", err, p)
	}
	_, err = p.Try(func(s downloadSource) (speedResult, error) {
		attempted = append(attempted, s.Key)
		return speedResult{Source: s.Name, Average: 2}, nil
	})
	if err != nil || strings.Join(attempted, ",") != "google,google_msi,google_msi" {
		t.Fatalf("unexpected options %v %v", attempted, err)
	}
}

// Custom external HTTPS file source through a real local HTTP CONNECT proxy,
// with a Range response and stable route-verification host.
func TestCustomFileHTTPSViaProxy(t *testing.T) {
	var seenRange atomic.Bool
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/files/custom.bin" {
			http.Error(w, "wrong URL", 404)
			return
		}
		if r.Header.Get("Range") == "bytes=0-999999" {
			seenRange.Store(true)
		} else {
			http.Error(w, "Range missing", 400)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", "1000000")
		w.WriteHeader(http.StatusPartialContent)
		_, _ = io.CopyN(w, strings.NewReader(strings.Repeat("Z", 1000000)), 1000000)
	}))
	defer srv.Close()
	proxy, port := localConnectProxy(t)
	defer proxy.Close()
	src, err := customDownloadSource(srv.URL + "/files/custom.bin")
	if err != nil {
		t.Fatal(err)
	}
	u, _ := url.Parse(srv.URL)
	got, err := benchmarkDownload(context.Background(), "Node", port, testProfile{Streams: 1, PerStreamMB: 1, Duration: 200 * time.Millisecond}, src, "test", testClientFactory(t, srv), func(ctx context.Context, host string, p int) (bool, bool, string) {
		if host != u.Hostname() || p <= 0 {
			t.Errorf("bad verifier host/port %q %d", host, p)
		}
		return true, true, "✓ Node"
	})
	if err != nil || got.Average <= 0 || !seenRange.Load() {
		t.Fatalf("custom benchmark %+v %v", got, err)
	}
}
func TestRedirectNotMisinterpretedAsDownload(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://example.net/other.bin", 302)
	}))
	defer srv.Close()
	out, err := runLocalBenchmark(t, srv, testProfile{Streams: 1, PerStreamMB: 1, Duration: 200 * time.Millisecond}, false)
	if err == nil || out.Average != 0 || !strings.Contains(err.Error(), "302") {
		t.Fatalf("unexpected redirect treatment %+v %v", out, err)
	}
}
