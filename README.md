# ClashBandwidthTest

一个面向 Windows 的轻量级 Clash Verge Rev / Mihomo 节点带宽测试与节点管理辅助工具。

它解决的核心问题很简单：Clash 自带的延迟测试只能告诉你 `ms`，但无法直观回答“这个节点到底能跑多少 MB/s”。ClashBandwidthTest 会自动连接本机 Mihomo，按地区/名称筛选节点，进行真实下载吞吐测试，并验证测速流量实际经过了被测节点。

> 当前版本：**v2.0.1**  
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
3. 程序会自动检测本机 Clash/Mihomo 配置；正常情况下无需手动填写 Controller 或 Secret。
4. 选择策略组和测速范围：
   - 地区智能抽样
   - 指定地区
   - 按名称筛选
   - 全部节点
5. 选择快速筛选或标准测速，开始测试。
6. 测速结束后，可选择单个节点“重测选中”，或双击节点直接切换。

## 结果字段

- **延迟**：节点延迟测试结果。
- **平均速度**：整个正式测速阶段的平均下载吞吐，主要参考指标。
- **稳态 P90**：短时间窗吞吐的 90 百分位，用于观察稳定状态下的高位表现。
- **峰值速度**：瞬时最高吞吐，仅作参考。
- **路径验证**：确认测速连接的代理链中包含当前被测节点，避免因规则分流到 DIRECT/其他节点导致假结果。

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

程序不会把 Clash Secret 或订阅信息写入自己的设置文件。

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
