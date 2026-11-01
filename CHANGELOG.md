## v2.0.9 (Static-file download source redesign)

### Changed
- Default auto source: Google Chrome static file -> Google Chrome Windows MSI -> Hetzner 100MB file. Cloudflare and OVH remain available as manual choices only.
- Added HTTPS custom file URL option; never persists the URL or accepts embedded URL credentials.
- Uses HTTP Range and local per-stream byte caps; rejects HTTP 200 HTML interstitial/error pages and redirects instead of measuring error content.
- Preserves single-source pinning within a run, precise Mihomo TCP source-port route verification, regional filtering, one-node-at-a-time workflow, local encrypted controller memory, and System-DPI awareness.

### Tests
- Regression tests for custom HTTPS file download via CONNECT proxy, source selection/pinning, HTML 200, HTTP 302, HTTP 429, early EOF, slow streaming, cancellations, Range handling.
- Windows x64 build, static analysis, race testing and artifact hashes checked locally. Does not prove external CDN availability or user-node performance.

## v2.0.8 — Rate-limit protection and single-stream quick benchmarking (REVIEWED)

### Fixed
- HTTP 429 from a public benchmark host is reported as origin rate limiting, never as a failed proxy or a comparable partial throughput score.
- The first real download stream now serves as the route verification probe, eliminating the extra HTTPS request per node.
- Fast mode now runs 1 stream (was 2), standard runs 2 (was 4): scores are not directly comparable to older multistream results.
- Batch runs stop after a pinned origin returns HTTP 429, marking untested nodes as untested; protect previously completed scores and restore the original proxy selection.
- Insert 2.5-second inter-node pacing to reduce pressure on public speed-test infrastructure. This cannot guarantee the origin will not impose rate limits.
- Keep DPAPI connection recovery, system DPI stability and route verification unchanged.

### Verification
- Local HTTP CONNECT + TLS tests for rate-limited first/second streams, no repeated 429 requests, no extra preflight request, byte caps, wrong-route refusal, partial transport failures and cancellation.
- Windows x64 cross-compilation, Go vet and race-detector tests; real OVH access and the user's Mihomo network remain unverified until user acceptance.

## v2.0.7 — Source selection and short-probe download fix (REVIEWED)

### Fixed
- Eliminate the full-megabyte pre-download that could time out before the 2-second benchmark started on 0.1 MB/s links. Use a bounded 16 KiB HTTPS data-and-route probe instead.
- Preserve a separately timed bandwidth stage; count only its bytes and elapsed measurement time.
- Auto-source preference is OVH Singapore, Hetzner Germany, Cloudflare. Pin the first successfully measured source for the entire batch; no silent cross-origin score mixing.
- Add manual Hetzner selection while retaining manual Cloudflare and OVH modes.
- Keep the byte caps, proxy-path verification, DPAPI controller recovery, single instance, tray behavior, and mixed-DPI stability.
- Make the estimate clear that automatic-source retries can add traffic on top of ordinary single-source usage.

### Verification
- HTTP CONNECT + TLS integration tests with actual byte transfer, including throttled ~0.1MB/s streams; static/dynamic paths, fallback selection, range cap, wrong-route refusal, cancellation.
- Go unit/integration tests and five repeated race-detector passes; Windows amd64 go vet and cross-compilation.
- Actual local Clash Verge/Mihomo proxy connections and performance over the user's subscription remain unverified in this container.

## v2.0.6 — Download source isolation and transport QA (candidate)

### Fixed
- Isolate warm-up retries and download-stream timeouts, so a stalled warm-up cannot exhaust the entire benchmark without a clear error.
- Fix per-stream byte caps: static servers ignoring HTTP Range can no longer cause the downloader to consume an entire large test file per connection.
- Add Cloudflare and an independently documented OVH Singapore HTTPS static-file download source, selectable for a whole batch.
- In Auto mode, fallback to OVH only before the first successful source has been pinned; once a source succeeds, keep it fixed for the remaining nodes to preserve comparability.
- Continue to refuse a confirmed wrong route rather than masking it by changing the test server.
- Keep DPAPI credential storage, UI scaling, single-instance protection and node restoration unchanged.

### Verification
- HTTP CONNECT and TLS integration tests for complete downloads, Range support/ignore, early EOF, partial streams, cancellation, fallback pinning and route mismatches.
- Linux Go tests with race detection, Windows amd64 go vet and cross-compilation.
- Real Windows + Mihomo + user-specific provider network remains to be verified; never claim a remote node is validated by simulated tests.

## v2.0.5 — Download resilience (candidate)

### Fixed
- Retry failed Cloudflare warm-up and individual download-stream connection attempts at most once per request, with bounded waits and no route-verification bypass.
- Treat an early EOF as a truncated warm-up, not as successful full transfer. If at least 100 KB was received and both warm-up attempts were incomplete, a benchmark may proceed but any reported throughput is explicitly marked **部分完成**.
- When fewer than the requested parallel streams connect, run with the available streams and clearly label the measured throughput as **部分完成** (not directly comparable with full-stream scores).
- Detect prematurely closed benchmark streams and avoid reporting a short, unreliable measurement as a valid full result.
- Preserve per-node route verification and all v2.0.4 controller/secret persistence behavior.

### Limitations
- This release does not repair an external network/Cloudflare outage. When all download requests fail, it reports failure and the transport error rather than inventing a speed.
- Actual Windows + Mihomo integration has not been run in the build environment; real-device acceptance remains outstanding.

## v2.0.4

### Fixed
- Prefer a previously user-confirmed, DPAPI-protected local HTTP controller before probing potentially stale runtime config files.
- Persist a successfully auto-detected local HTTP controller when the user has explicitly enabled the remember option.
- Pre-fill the remembered controller and secret after auto-connect failure so the user does not need to retype credentials.
- Surface connection-persistence errors instead of silently hiding them.

### Improved
- Add a per-node error detail column for bandwidth-test failures.
- Report Cloudflare warm-up request/read errors, HTTP status and short response hints, individual stream-open failures, and actual received bytes when throughput is below the minimum threshold.
- Show success/failure counts after a test run or manual stop.
- Keep the existing measurement source and throughput algorithm unchanged pending diagnosis from the newly exposed error details.

### Security
- Continue to store remembered local controller credentials only through Windows current-user DPAPI; no plaintext Secret is added to settings.json.

## v2.0.3

### Fixed
- Prevent selecting GLOBAL for download benchmarking when Mihomo is running in Rule/Direct mode; this is a routing-mode mismatch, not a node-bandwidth failure.
- Prefer non-GLOBAL selectors by default and remember the user's last selected group.
- Detect repeated mismatches during one run and stop early with an actionable explanation.
- Remember an explicitly approved local HTTP controller and secret across restarts using Windows DPAPI (per-user encrypted file; opt-in checkbox in Advanced settings).
- Reuse remembered credentials when runtime configuration discovery fails.
- Preserve bandwidth algorithm, multi-monitor and tray behavior from v2.0.2.

## v2.0.2 (Route verification correction)

### Fixed
- Match the benchmark connection to Mihomo `/connections` by the exact local TCP source port instead of accepting the first new connection sharing the benchmark hostname.
- Support Mihomo `sourcePort` values serialized as strings or numbers.
- Report missing or unobservable connections as unverified, not as proven misrouting.
- Preserve per-stream isolated transports, disable keep-alive/HTTP2, and refuse redirects to other benchmark hosts.

### Verification
- Added focused unit tests for source-port parsing and attribution under concurrent connections.
- Windows runtime + Mihomo integration still requires actual device verification.

## v2.0.1

### Fixed
- Reworked single-instance detection using a session-local named mutex and reliable CreateMutexW error handling.
- Launching the EXE again now wakes the existing window instead of creating another process.
- Removed the WM_SETREDRAW/DPI rebuild workaround that could leave the ListView body blank.
- Switched to stability-first System DPI awareness so Windows scales the complete UI atomically between monitors.
- Centralized the visible application version; window title and tray tooltip now show 2.0.1.

### Note
- System DPI awareness favors stability across mixed-DPI displays. On a secondary monitor with a different scale factor, Windows may bitmap-scale the app slightly, which can look a little softer than Per-Monitor V2.

## v1.9.1a

### Fixed
- Fixed Go unsafe.Pointer build failure in WM_DPICHANGED handling.

## v1.9.1

### Fixed
- Properly apply Windows WM_DPICHANGED suggested window bounds before refreshing the UI.
- Improved ListView stability when moving between monitors with different DPI scaling.

## v1.9.0

### Fixed
- Added explicit DPI change handling for multi-monitor Windows setups.
- Improved ListView rebuild after DPI/layout changes.

## v1.8.4

### Fixed
- Improved ListView refresh behavior during window resizing.
- Prevented table content disappearing after resize/maximize operations.
- Improved stability when changing window size on multi-monitor setups.

# Changelog

## v1.8
- 新增系统托盘支持。
- 新增“关闭按钮缩小到托盘”。
- 新增启动时隐藏到托盘选项。
- 新增全局快捷键（默认 Ctrl+Alt+B）。
- 新增快捷键冲突检测与设置持久化。

## v1.7
- 新增双击节点直接切换。
- 新增“重测选中”。
- 修正切换节点后的当前节点状态记录。

## v1.6
- 新增“指定地区”与多地区选择。
- 新增按名称筛选。
- 支持地区内全部/抽样测试。

## v1.5
- 新增地区自动识别与智能抽样。
- 按线路类别优先覆盖不同候选。
- 新增两阶段延迟筛选 + 带宽精测流程。
- 新增异常退出恢复机制。
- 改进测速状态机与进度显示。

## v1.4
- 嵌入 Windows ICO 图标与 Manifest。
- 引入预热、多路并发与固定时间测速。
- 新增平均速度、稳态 P90、峰值速度。

## v1.3
- 新增路径验证，检查测速流量真实代理链。
- 新增预计测速流量提示。

## v1.2
- 自动发现 Clash Verge Rev / Mihomo。
- 自动读取本机配置、Mixed Port 与控制信息。
- 优先使用 Windows Named Pipe。

## v1.1
- 修复 Win32 GUI 消息循环线程导致的“未响应”问题。

## v1.0
- 初始版本：节点读取、延迟测试、下载带宽测速与原节点恢复。


## v1.8.2

### Added
- Single-instance protection to prevent multiple ClashBandwidthTest windows running simultaneously.

### Improved
- Prevent duplicate Clash controller operations caused by multiple instances.
