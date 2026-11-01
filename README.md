# ClashBandwidthTest

一个面向 Windows 的轻量级 Clash Verge Rev / Mihomo 节点带宽测试与节点管理辅助工具。

它解决的核心问题很简单：Clash 自带的延迟测试只能告诉你 `ms`，但无法直观回答“这个节点到底能跑多少 MB/s”。ClashBandwidthTest 会自动连接本机 Mihomo，按地区/名称筛选节点，进行真实下载吞吐测试，并验证测速流量实际经过了被测节点。

> 当前版本：**v2.0.9**  
> 平台：Windows x64  
> 语言：Go + Win32 API

## 功能

- 自动发现 Clash Verge Rev / Mihomo
- 支持地区智能抽样、指定地区、名称筛选、全部节点测速
- 延迟、平均下载速度、稳态 P90、峰值速度
- 路径验证：检查测速流量是否真正经过被测节点
- 快速筛选与标准测速模式
- 双击节点直接切换到该节点
- 仅重测选中节点
- 测速完成/停止后恢复原节点
- 异常退出恢复记录
- 系统托盘常驻
- 全局快捷键显示/隐藏主窗口（默认 `Ctrl+Alt+B`）
- 可配置“关闭按钮最小化到托盘”“启动时最小化到托盘”等行为

## 使用方式

1. 启动 Clash Verge Rev，并确保 Mihomo 核心正常运行。
2. 启动 ClashBandwidthTest。
3. 程序会优先尝试上次明确确认并经 DPAPI 加密保存的本机 Controller，再读取 Clash Verge Rev 运行配置；正常情况下无需每次手动填写 Controller 或 Secret。
4. 选择策略组和测速范围：
   - 地区智能抽样
   - 指定地区
   - 按名称筛选
   - 全部节点
5. 选择快速筛选或标准测速，设置测速源（默认“自动（Google优先）”），开始测试。
   - 默认改为下载**大型静态文件**，自动顺序：**Google Chrome 文件 → Google Chrome Windows 文件 → Hetzner 德国**；不再把频繁触发 429 的 Cloudflare / OVH 放入自动轮询。
   - 可以手动选择 Google、Hetzner、OVH、Cloudflare，或选择**自定义 HTTPS 文件**。选择自定义后在“高级”下方填入完整文件下载 URL（请勿填写有 Token 或凭据的链接）。
   - 使用静态文件时，发送 `Range: bytes=0-N`，即使服务器忽略 Range，客户端也限制最多接收 N 字节；返回 HTML 错误页或 3xx 跳转时不会误当成速度。
   - 第一个**真正完成测速**的来源将在本轮锁定，所有正常结果来自同一下载地址；一旦检测到可信的错误路由，则直接停止回退，避免混淆。
   - 从 v2.0.8 起，第一条正式下载连接同时完成代理路径验证，不再额外发送 16 KiB/1 MB 预检请求。计时只统计正式下载的读取阶段。
   - **注意流量**：页面估算是单一测速源正常完成时的上限；自动回退与失败重试可能额外消耗流量；单条连接独立设置字节上限。
6. 测速结束后，可选择单个节点“重测选中”，或双击节点直接切换。

## 结果字段

- **延迟**：节点延迟测试结果。
- **平均速度**：整个正式测速阶段的平均下载吞吐，主要参考指标。
- **稳态 P90**：短时间窗吞吐的 90 百分位，反映观测窗口内较高的吞吐，不等同于严格的稳态测量。
- **峰值速度**：瞬时最高吞吐，仅作参考。
- **路径验证**：确认测速连接的代理链中包含当前被测节点，避免因规则分流到 DIRECT/其他节点导致假结果。
- **错误详情**：包含所用测速源、重试/截断说明及失败原因。
- **部分完成**：并发流不完整或下载提前中断，测速值仅供参考，不应与完整测速直接排序比较。

## 托盘与快捷键

默认行为：

- 点击主窗口右上角 `×`：隐藏到系统托盘，不退出程序。
- 双击托盘图标：恢复主窗口。
- 托盘右键：显示主窗口 / 设置 / 退出。
- 默认全局快捷键：`Ctrl+Alt+B`。
- 选择托盘菜单“退出”才会真正退出；如果正在测速，会先停止并尝试恢复测速前节点。

设置保存在：

```text
%APPDATA%\ClashBandwidthTest\settings.json
```

程序不会把 Clash Secret 或订阅信息写入自己的普通设置文件。若明确勾选“记住连接”，仅本机 Controller 和 Secret 会使用 Windows 当前用户 DPAPI 加密保存。

## 网络设置说明

ClashBandwidthTest **不会修改** Windows System Proxy、TUN、DNS、WFP、路由表、网卡或注册表代理设置。

测速期间，它只会通过 Mihomo API 临时切换用户选择的 Selector 策略组中的节点；正常结束、停止或正常退出时会恢复测速前节点。

测速时，正在使用同一 Selector 的其他程序可能因为节点切换而短暂断线。

## 构建

需要 Go 1.23+。

仓库中已包含 Windows 资源文件 `rsrc_windows_amd64.syso`，在 Windows amd64 环境下可直接：

```powershell
go build -trimpath -ldflags="-s -w -H=windowsgui" -o ClashBandwidthTest.exe .
```

或在 Linux/macOS 上交叉编译：

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
  go build -trimpath -ldflags="-s -w -H=windowsgui" -o ClashBandwidthTest.exe .
```

## 安全提示

当前发行版未进行商业代码签名，因此 Windows SmartScreen 可能显示“未知发布者”。建议从本仓库源码自行构建，或核对 Release 中提供的 SHA-256。

## 项目状态

这是一个个人使用场景驱动的小型项目，目前重点是稳定性、测速可信度与使用体验。欢迎提交 Issue 描述可复现的问题或功能建议。

## 致谢

本项目依赖 Mihomo 提供的本地控制 API，并面向 Clash Verge Rev / Mihomo 使用场景。项目与 Clash Verge Rev、Mihomo 官方项目无隶属关系。

## v2.0.4: 连接恢复与测速失败诊断

- 自动连接优先验证上次手动确认的 Controller，避免被过期运行配置中的地址抢先匹配。
- 勾选“记住连接”后，自动检测成功的本机 HTTP Controller 也会更新加密记忆；自动连接失败时会回填已保存的地址和 Secret，减少重复输入。
- 结果表新增“错误详情”，显示 Cloudflare 预热、HTTP 状态、并发下载连接和实际接收字节数等失败原因。
- 完成或停止后显示成功/失败数量；不改变原有延迟与带宽测量策略，不自动切换测速源。

## v2.0.3: 连接记忆与 Rule 模式

- 当 Clash 处于 Rule 模式时，不要用 `GLOBAL` 作为带宽测速策略组；应选择实际路由使用的 Selector。工具会在开始测速前检查。
- 若自动发现失败，可在“高级”中输入本机 Controller 和 Secret，勾选 **记住连接（本机加密）** 再点击“手动连接”。下一次启动自动尝试本机记忆连接。
- 记忆文件位于 `%LOCALAPPDATA%\ClashBandwidthTest\controller.dpapi`，使用 Windows 当前用户 DPAPI 加密。此加密不能防御以同一 Windows 用户身份运行的恶意程序。取消勾选并成功手动连接，会删除之前保存的记忆文件。
- 不会上传 Secret，也不会在 `settings.json` 中以明文保存 Secret。

## v2.0.5: Cloudflare 连接中断处理

- 预热和下载流建连阶段支持**最多一次**有限重试，不会无限占用节点流量。
- 若预热获得足够数据但中途 EOF，会标记该次预热不完整，继续测速时显示警告。
- 双路/四路测试若只有部分连接成功，提供明确的“部分完成”结果；**不可与满并发流的成绩直接比较**。
- 若 Cloudflare 或中间代理持续拒绝连接，仍会显示“测速失败”，不会捏造速度。
- 保留已通过用户实机初验的 v2.0.4 自动连接、Secret 加密记忆、分流路径验证、单实例和 DPI 设置。

如需区分程序问题与节点到 Cloudflare 的网络问题，可在已选目标节点下运行：

```powershell
curl.exe --proxy http://127.0.0.1:7897 --max-time 20 -o NUL -w "HTTP=%{http_code} bytes=%{size_download} time=%{time_total}\n" "https://speed.cloudflare.com/__down?bytes=1000000"
```

仅当 Mixed Port 仍是 7897 时直接使用，否则替换为 Clash 当前 Mixed Port。

## v2.0.7 历史版本验证范围

- Go 测试覆盖 HTTP CONNECT 代理 + HTTPS 实际下载、约 0.1MB/s 的限速网络模拟、16 KiB 预检、Cloudflare 动态下载、OVH/Hetzner 静态文件 Range 下载、忽略 Range 时字节截断、unexpected EOF、单流失败、超时取消、三源降级、同批固定来源以及错误分流拒绝。
- 本地测试不能复制真实节点/运营商路径、第三方测速源的可用性或 Windows 桌面行为；真正的全球节点速度仍须在使用者环境下验收。
- 本工具不得自动更改用户的系统代理/TUN/DNS 设置，也不上传本地 Secret。


## v2.0.8 逐节点测速与测速源限流保护

“指定地区（全部）”“按名称筛选”“全部节点”模式：逐节点测延迟、立即测下载，不需要先对全体节点执行一遍延迟扫描。只有智能抽样或指定地区抽样时先筛延迟再精测。

快速模式为 **单连接 2 秒**；标准模式为 **双连接 5 秒**。旧版更多并发流的下载速度不能直接与此版本横向比较。每节点仅创建正式测速连接，不额外预检。

公共测速服务器可能返回 HTTP 429（请求过多）。当已经固定的测速源返回 429 时，本轮立即停止，后续节点标为“未测（测速源限流）”，不算作节点测速失败；测速源尚未固定时自动模式可继续尝试备用源。节点之间增加 2.5 秒间隔以降低请求密度，但任何公共服务都无法保证无限制自动批量测速。

## v2.0.9: 静态下载测速修复

- 参考 [faceair/clash-speedtest](https://github.com/faceair/clash-speedtest) 的**设计思路**（该项目 GPL-3.0）；本项目未复制其源代码。
- 下载测速默认使用 Google Chrome 大文件，支持 Windows MSI 文件与 Hetzner 备选；Cloudflare 和 OVH 仅供手动选取，避免批量任务默认触发 429 限流。
- 支持在“高级”输入 HTTPS 静态文件 URL；该 URL 不保存到本机设置或 GitHub。
- Range 与最大字节上限独立控制；拒绝 HTTP 200 的 HTML 错误页、无效下载响应，不将这些数据纳入测速。
- 已通过代理 TLS 集成测试及 Windows x64 交叉编译；真实订阅节点与外部 Google CDN 的可用性需实际设备验收。
