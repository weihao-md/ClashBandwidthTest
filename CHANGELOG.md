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
