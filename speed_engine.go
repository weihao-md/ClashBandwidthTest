package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type testProfile struct {
	Source      string
	CustomURL   string
	Name        string
	WarmupMB    int // legacy profile field; no longer a full 1MB pre-transfer
	Streams     int
	PerStreamMB int
	Duration    time.Duration
	Settle      time.Duration
}

func (p testProfile) MaxMBPerNode() int {
	return p.Streams * p.PerStreamMB // per-node cap; no separate download preflight
}

type speedResult struct {
	Source     string
	Warning    string
	Average    float64
	Peak       float64
	P90        float64
	Route      string
	RouteFound bool
	RouteOK    bool
}

type openedStream struct {
	body io.ReadCloser
	tr   *http.Transport
}

func newSpeedClient(proxyPort int) (*http.Client, *http.Transport, *atomic.Int64) {
	proxyURL, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", proxyPort))
	sourcePort := &atomic.Int64{}
	dialer := &net.Dialer{Timeout: 8 * time.Second}
	tr := &http.Transport{
		Proxy:                 http.ProxyURL(proxyURL),
		DisableKeepAlives:     true,
		DisableCompression:    true,
		TLSHandshakeTimeout:   8 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ForceAttemptHTTP2:     false,
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			conn, err := dialer.DialContext(ctx, network, address)
			if err != nil {
				return nil, err
			}
			if a, ok := conn.LocalAddr().(*net.TCPAddr); ok {
				sourcePort.Store(int64(a.Port))
			}
			return conn, nil
		},
	}
	// Benchmark requests must never silently change host through a redirect.
	client := &http.Client{Transport: tr, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	return client, tr, sourcePort
}

func percentile90(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	v := append([]float64(nil), values...)
	sort.Float64s(v)
	idx := (9*len(v)+9)/10 - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(v) {
		idx = len(v) - 1
	}
	return v[idx]
}

// SourceHTTPError preserves status semantics so a public speed-test server's
// 429 response is never misreported as a broken node or a partial speed score.
type SourceHTTPError struct {
	Source string
	Code   int
	Stage  string
}

func (e *SourceHTTPError) Error() string {
	if e.Code == http.StatusTooManyRequests {
		return fmt.Sprintf("%s：HTTP 429（测速服务器限流）；%s；请停止本轮测试并稍后使用其他测速源", e.Source, e.Stage)
	}
	return fmt.Sprintf("%s：%s HTTP %d", e.Source, e.Stage, e.Code)
}
func isRateLimited(err error) bool {
	var he *SourceHTTPError
	return errors.As(err, &he) && he.Code == http.StatusTooManyRequests
}

// benchmarkDownload uses the first *real measured* HTTPS download connection
// to inspect the Mihomo route. There is no redundant warm-up HTTP request per
// node. The batch driver will stop on HTTP 429 rather than hammering the origin.
func benchmarkDownload(ctx context.Context, node string, proxyPort int, profile testProfile, source downloadSource, version string,
	clientFactory func(int) (*http.Client, *http.Transport, *atomic.Int64), verify func(context.Context, string, int) (bool, bool, string)) (speedResult, error) {
	var out speedResult
	out.Source = source.Name
	if profile.Streams < 1 || profile.PerStreamMB <= 0 || profile.Duration <= 0 {
		return out, fmt.Errorf("无效的测速参数")
	}
	streamBytes := int64(profile.PerStreamMB) * 1000 * 1000
	openStream := func(parent context.Context, streamNo int) (openedStream, int, error) {
		var lastErr error
		for attempt := 1; attempt <= 2; attempt++ {
			if err := parent.Err(); err != nil {
				return openedStream{}, 0, err
			}
			// Opening may take 12 s, but a successful response must retain its
			// request context until the body is read. Canceling at Do() return
			// would reliably kill the HTTPS download immediately.
			requestCtx, requestCancel := context.WithCancel(parent)
			openTimer := time.AfterFunc(12*time.Second, requestCancel)
			client, tr, local := clientFactory(proxyPort)
			req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, source.URL(streamBytes), nil)
			if err != nil {
				openTimer.Stop()
				requestCancel()
				tr.CloseIdleConnections()
				return openedStream{}, 0, err
			}
			req.Header.Set("User-Agent", "ClashBandwidthTest/"+version)
			req.Header.Set("Cache-Control", "no-store")
			req.Header.Set("Accept-Encoding", "identity")
			if source.SupportsRange {
				req.Header.Set("Range", fmt.Sprintf("bytes=0-%d", streamBytes-1))
			}
			resp, e := client.Do(req)
			openTimer.Stop()
			if e == nil && (resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusPartialContent) {
				// A captive portal / CDN error page can return HTTP 200. Do
				// not silently measure HTML as if it were a binary download.
				if media := strings.ToLower(resp.Header.Get("Content-Type")); strings.HasPrefix(media, "text/html") || strings.HasPrefix(media, "application/xhtml+xml") {
					resp.Body.Close()
					requestCancel()
					tr.CloseIdleConnections()
					return openedStream{}, 0, fmt.Errorf("%s：返回 HTML 页面而不是下载文件", source.Name)
				}
				return openedStream{body: &cancelOnCloseReader{ReadCloser: resp.Body, cancel: requestCancel}, tr: tr}, int(local.Load()), nil
			}
			requestCancel()
			tr.CloseIdleConnections()
			if e == nil {
				he := &SourceHTTPError{Source: source.Name, Code: resp.StatusCode, Stage: fmt.Sprintf("第 %d 路下载", streamNo)}
				resp.Body.Close()
				lastErr = he
				// 429 and all 4xx are deterministic for a specific URL, never retry.
				if resp.StatusCode >= 400 && resp.StatusCode < 500 {
					return openedStream{}, 0, he
				}
			} else {
				lastErr = fmt.Errorf("%s：第 %d 路建立下载连接失败（%d/2）：%w", source.Name, streamNo, attempt, e)
			}
			if attempt < 2 {
				select {
				case <-parent.Done():
					return openedStream{}, 0, parent.Err()
				case <-time.After(250 * time.Millisecond):
				}
			}
		}
		return openedStream{}, 0, lastErr
	}

	// One shared cancellation context owns every measured stream, including the
	// first. The 2/5-second timer must interrupt a blocked first stream read.
	benchCtx, benchCancel := context.WithCancel(ctx)
	defer benchCancel()
	// Start the first stream and verify its exact local TCP source port while
	// its response stays open. This replaces the extra 16-KB warmup request.
	first, localPort, err := openStream(benchCtx, 1)
	if err != nil {
		return out, err
	}
	verificationCtx, verificationCancel := context.WithTimeout(ctx, 6*time.Second)
	found, routeOK, routeDesc := verify(verificationCtx, source.Host, localPort)
	verificationCancel()
	out.RouteFound, out.RouteOK, out.Route = found, routeOK, routeDesc
	if found && !routeOK {
		first.body.Close()
		first.tr.CloseIdleConnections()
		return out, fmt.Errorf("测速流量没有经过当前节点：%s", routeDesc)
	}
	if ctx.Err() != nil {
		first.body.Close()
		first.tr.CloseIdleConnections()
		return out, ctx.Err()
	}

	// Subsequent streams are only opened after the first route was verified.
	streams := []openedStream{first}
	type openResult struct {
		s   openedStream
		err error
	}
	openCh := make(chan openResult, profile.Streams-1)
	for i := 1; i < profile.Streams; i++ {
		no := i + 1
		go func() {
			st, _, e := openStream(benchCtx, no)
			if e != nil {
				openCh <- openResult{err: e}
				return
			}
			openCh <- openResult{s: st}
		}()
	}
	var firstOpenErr error
	var limitErr error
	for i := 1; i < profile.Streams; i++ {
		r := <-openCh
		if r.err != nil {
			if firstOpenErr == nil {
				firstOpenErr = r.err
			}
			if isRateLimited(r.err) {
				limitErr = r.err
			}
			continue
		}
		streams = append(streams, r.s)
	}
	if limitErr != nil {
		for _, st := range streams {
			st.body.Close()
			st.tr.CloseIdleConnections()
		}
		return out, limitErr
	}
	if len(streams) < profile.Streams {
		out.Warning = fmt.Sprintf("部分完成：仅 %d/%d 条并发流建立成功", len(streams), profile.Streams)
		if firstOpenErr != nil {
			out.Warning += "；" + firstOpenErr.Error()
		}
	}

	// Only time actual body reads. Each client stream has a byte cap even if
	// a static server ignores Range and responds with a large entire file.
	var total int64
	var truncated atomic.Bool
	var readers sync.WaitGroup
	readers.Add(len(streams))
	for _, st := range streams {
		st := st
		go func() {
			defer readers.Done()
			defer st.body.Close()
			defer st.tr.CloseIdleConnections()
			buf := make([]byte, 128*1024)
			var got int64
			for {
				if benchCtx.Err() != nil || got >= streamBytes {
					return
				}
				maxN := len(buf)
				if remain := streamBytes - got; remain < int64(maxN) {
					maxN = int(remain)
				}
				n, er := st.body.Read(buf[:maxN])
				if n > 0 {
					got += int64(n)
					atomic.AddInt64(&total, int64(n))
				}
				if er != nil {
					if got < streamBytes && benchCtx.Err() == nil {
						truncated.Store(true)
					}
					return
				}
			}
		}()
	}
	allDone := make(chan struct{})
	go func() { readers.Wait(); close(allDone) }()
	start := time.Now()
	lastAt := start
	var lastBytes int64
	var samples []float64
	peak := 0.0
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	timer := time.NewTimer(profile.Duration)
	defer timer.Stop()
loop:
	for {
		select {
		case now := <-ticker.C:
			cur := atomic.LoadInt64(&total)
			dt := now.Sub(lastAt).Seconds()
			if dt > 0 {
				rate := float64(cur-lastBytes) / 1000000.0 / dt
				if rate >= 0 {
					samples = append(samples, rate)
					if rate > peak {
						peak = rate
					}
				}
			}
			lastBytes, lastAt = cur, now
		case <-allDone:
			break loop
		case <-timer.C:
			benchCancel()
			break loop
		case <-ctx.Done():
			benchCancel()
			readers.Wait()
			return out, ctx.Err()
		}
	}
	benchCancel()
	readers.Wait()
	elapsed := time.Since(start).Seconds()
	bytesDone := atomic.LoadInt64(&total)
	if truncated.Load() {
		if out.Warning != "" {
			out.Warning += "；"
		}
		out.Warning += "部分下载流提前中断，结果仅供参考"
	}
	if truncated.Load() && elapsed < 0.5 && bytesDone < int64(len(streams))*streamBytes {
		return out, fmt.Errorf("下载流提前中断，观测不足 %.2f 秒（%d 字节）", elapsed, bytesDone)
	}
	if bytesDone < 16384 || (bytesDone < 100000 && elapsed < 1.0) {
		return out, fmt.Errorf("有效下载数据不足：%d 字节，耗时 %.2f 秒", bytesDone, elapsed)
	}
	if bytesDone < 100000 {
		if out.Warning != "" {
			out.Warning += "；"
		}
		out.Warning += "低吞吐小样本，仅供参考"
	}
	if elapsed <= 0 {
		return out, fmt.Errorf("无效计时")
	}
	avg := float64(bytesDone) / 1000000.0 / elapsed
	if peak <= 0 {
		peak = avg
	}
	p90 := percentile90(samples)
	if p90 <= 0 {
		p90 = avg
	}
	out.Average, out.Peak, out.P90 = avg, peak, p90
	return out, nil
}

type cancelOnCloseReader struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (c *cancelOnCloseReader) Close() error {
	err := c.ReadCloser.Close()
	c.cancel()
	return err
}
