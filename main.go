//go:build windows

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

const (
	appTitle = "ClashBandwidthTest"

	IDC_CONTROLLER  = 1001
	IDC_SECRET      = 1002
	IDC_CONNECT     = 1003
	IDC_GROUP       = 1004
	IDC_SIZE        = 1005
	IDC_START       = 1006
	IDC_STOP        = 1007
	IDC_LIST        = 1008
	IDC_STATUS      = 1009
	IDC_ADVANCED    = 1010
	IDC_REDETECT    = 1011
	IDC_ESTIMATE    = 1012
	IDC_SCOPE       = 1013
	IDC_PERREGION   = 1014
	IDC_REGIONS     = 1015
	IDC_FILTER      = 1016
	IDC_SCOPEDETAIL = 1017
	IDC_RETEST      = 1018
	IDC_SETTINGS    = 1019

	IDC_REGION_OK     = 2101
	IDC_REGION_CANCEL = 2102
	IDC_REGION_BASE   = 2200

	IDC_SET_CLOSETRAY  = 3101
	IDC_SET_STARTTRAY  = 3102
	IDC_SET_HOTKEY     = 3103
	IDC_SET_HOTKEYEDIT = 3104
	IDC_SET_SAVE       = 3105
	IDC_SET_CANCEL     = 3106

	ID_TRAY_SHOW     = 4101
	ID_TRAY_SETTINGS = 4102
	ID_TRAY_EXIT     = 4103
	HOTKEY_ID        = 4201
	TRAY_ICON_ID     = 1

	WM_APP_REFRESH = 0x8001
	WM_APP_GROUPS  = 0x8002
	WM_APP_DONE    = 0x8003
	WM_APP_DETECT  = 0x8004
	WM_APP_STATUS  = 0x8005
	WM_APP_RECOVER = 0x8006
	WM_APP_SWITCH  = 0x8007
	WM_APP_LISTDBL = 0x8008
	WM_APP_TRAY    = 0x8009
	WM_APP_WAKE    = 0x8010

	WS_OVERLAPPED     = 0x00000000
	WS_CAPTION        = 0x00C00000
	WS_SYSMENU        = 0x00080000
	WS_MINIMIZEBOX    = 0x00020000
	WS_VISIBLE        = 0x10000000
	WS_CHILD          = 0x40000000
	WS_BORDER         = 0x00800000
	WS_TABSTOP        = 0x00010000
	WS_VSCROLL        = 0x00200000
	ES_PASSWORD       = 0x0020
	BS_PUSHBUTTON     = 0x00000000
	BS_AUTOCHECKBOX   = 0x00000003
	CBS_DROPDOWNLIST  = 0x0003
	LVS_REPORT        = 0x0001
	LVS_SINGLESEL     = 0x0004
	LVS_SHOWSELALWAYS = 0x0008
	WS_EX_CLIENTEDGE  = 0x00000200

	CW_USEDEFAULT uintptr = 0x80000000
	SW_HIDE               = 0
	SW_SHOW               = 5
	SW_RESTORE            = 9
	SWP_NOZORDER          = 0x0004
	SWP_NOACTIVATE        = 0x0010

	WM_DESTROY       = 0x0002
	WM_SIZE          = 0x0005
	WM_DPICHANGED    = 0x02E0
	WM_ENTERSIZEMOVE  = 0x0231
	WM_EXITSIZEMOVE   = 0x0232
	WM_SETREDRAW      = 0x000B
	WM_COMMAND       = 0x0111
	WM_CLOSE         = 0x0010
	WM_SETFONT       = 0x0030
	WM_HOTKEY        = 0x0312
	WM_LBUTTONDBLCLK = 0x0203
	WM_RBUTTONUP     = 0x0205

	CB_ADDSTRING    = 0x0143
	CB_RESETCONTENT = 0x014B
	CB_SETCURSEL    = 0x014E
	CB_GETCURSEL    = 0x0147
	CBN_SELCHANGE   = 1

	EM_SETPASSWORDCHAR = 0x00CC
	EN_CHANGE          = 0x0300

	BM_GETCHECK = 0x00F0
	BM_SETCHECK = 0x00F1
	BST_CHECKED = 1

	LVM_FIRST                    = 0x1000
	LVM_INSERTITEMW              = LVM_FIRST + 77
	LVM_SETITEMTEXTW             = LVM_FIRST + 116
	LVM_INSERTCOLUMNW            = LVM_FIRST + 97
	LVM_SETCOLUMNWIDTH           = LVM_FIRST + 30
	LVM_DELETEALLITEMS           = LVM_FIRST + 9
	LVM_SETEXTENDEDLISTVIEWSTYLE = LVM_FIRST + 54
	LVM_GETNEXTITEM              = LVM_FIRST + 12

	LVS_EX_FULLROWSELECT = 0x00000020
	LVS_EX_GRIDLINES     = 0x00000001
	LVS_EX_DOUBLEBUFFER  = 0x00010000

	LVCF_FMT    = 0x0001
	LVCF_WIDTH  = 0x0002
	LVCF_TEXT   = 0x0004
	LVCFMT_LEFT = 0x0000

	LVIF_TEXT     = 0x0001
	LVNI_SELECTED = 0x0002
	NM_DBLCLK     = -3

	ICC_LISTVIEW_CLASSES = 0x00000001

	MB_OK              = 0x00000000
	MB_ICONERROR       = 0x00000010
	MB_ICONWARNING     = 0x00000030
	MB_ICONINFORMATION = 0x00000040
	MB_YESNO           = 0x00000004
	IDYES              = 6

	NIM_ADD              = 0x00000000
	NIM_MODIFY           = 0x00000001
	NIM_DELETE           = 0x00000002
	NIM_SETVERSION       = 0x00000004
	NIF_MESSAGE          = 0x00000001
	NIF_ICON             = 0x00000002
	NIF_TIP              = 0x00000004
	NOTIFYICON_VERSION_4 = 4

	MF_STRING       = 0x00000000
	MF_SEPARATOR    = 0x00000800
	TPM_RIGHTBUTTON = 0x0002
	TPM_RETURNCMD   = 0x0100

	MOD_ALT      = 0x0001
	MOD_CONTROL  = 0x0002
	MOD_SHIFT    = 0x0004
	MOD_WIN      = 0x0008
	MOD_NOREPEAT = 0x4000

	IDC_ARROW        = 32512
	COLOR_WINDOW     = 5
	DEFAULT_GUI_FONT = 17
	GWLP_WNDPROC     = -4
)

type WNDCLASSEX struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     syscall.Handle
	HIcon         syscall.Handle
	HCursor       syscall.Handle
	HbrBackground syscall.Handle
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       syscall.Handle
}

type MSG struct {
	HWnd    syscall.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type POINT struct{ X, Y int32 }
type RECT struct { Left, Top, Right, Bottom int32 }

type INITCOMMONCONTROLSEX struct{ DwSize, DwICC uint32 }

type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type NOTIFYICONDATA struct {
	CbSize            uint32
	HWnd              syscall.Handle
	UID               uint32
	UFlags            uint32
	UCallbackMessage  uint32
	HIcon             syscall.Handle
	SzTip             [128]uint16
	DwState           uint32
	DwStateMask       uint32
	SzInfo            [256]uint16
	UTimeoutOrVersion uint32
	SzInfoTitle       [64]uint16
	DwInfoFlags       uint32
	GuidItem          GUID
	HBalloonIcon      syscall.Handle
}

type appSettings struct {
	CloseToTray    bool   `json:"close_to_tray"`
	StartMinimized bool   `json:"start_minimized"`
	HotkeyEnabled  bool   `json:"hotkey_enabled"`
	Hotkey         string `json:"hotkey"`
}

type LVCOLUMN struct {
	Mask       uint32
	Fmt        int32
	Cx         int32
	PszText    *uint16
	CchTextMax int32
	ISubItem   int32
	IImage     int32
	IOrder     int32
	CxMin      int32
	CxDefault  int32
	CxIdeal    int32
}

type LVITEM struct {
	Mask       uint32
	IItem      int32
	ISubItem   int32
	State      uint32
	StateMask  uint32
	PszText    *uint16
	CchTextMax int32
	IImage     int32
	LParam     uintptr
	IIndent    int32
	IGroupId   int32
	CColumns   uint32
	PuColumns  *uint32
	PiColFmt   *int32
	IGroup     int32
}

type proxyInfo struct {
	Type string   `json:"type"`
	Now  string   `json:"now"`
	All  []string `json:"all"`
}

type proxiesResponse struct {
	Proxies map[string]proxyInfo `json:"proxies"`
}

type configResponse struct {
	MixedPort int `json:"mixed-port"`
	Port      int `json:"port"`
}

// connectionSnapshot mirrors the subset of Mihomo /connections used to verify
// that the speed-test request really traverses the node selected in the group.
type connectionSnapshot struct {
	Connections []struct {
		ID       string   `json:"id"`
		Download int64    `json:"download"`
		Chains   []string `json:"chains"`
		Rule     string   `json:"rule"`
		Metadata struct {
			Host          string `json:"host"`
			DestinationIP string `json:"destinationIP"`
		} `json:"metadata"`
	} `json:"connections"`
}

type resultRow struct {
	Name    string
	Region  string
	Line    string
	Type    string
	Latency int
	Speed   float64
	Peak    float64
	P90     float64
	Route   string
	Status  string
}

type clashEndpoint struct {
	Base   string
	Secret string
	Pipe   string
	Mode   string
}

type detectedConfig struct {
	Path       string
	Controller string
	Secret     string
	Pipe       string
	MixedPort  int
}

type pipeAddr string

func (a pipeAddr) Network() string { return "npipe" }
func (a pipeAddr) String() string  { return string(a) }

type namedPipeConn struct {
	h syscall.Handle
}

func (c *namedPipeConn) Read(p []byte) (int, error) {
	var n uint32
	err := syscall.ReadFile(c.h, p, &n, nil)
	if err != nil {
		// Message-mode named pipes may return ERROR_MORE_DATA together with a
		// valid partial read. net/http can continue reading the remainder.
		if err == syscall.ERROR_MORE_DATA && n > 0 {
			return int(n), nil
		}
		if err == syscall.ERROR_BROKEN_PIPE {
			return 0, io.EOF
		}
		return int(n), err
	}
	return int(n), nil
}

func (c *namedPipeConn) Write(p []byte) (int, error) {
	var n uint32
	err := syscall.WriteFile(c.h, p, &n, nil)
	if err != nil {
		return int(n), err
	}
	return int(n), nil
}

func (c *namedPipeConn) Close() error                     { return syscall.CloseHandle(c.h) }
func (c *namedPipeConn) LocalAddr() net.Addr              { return pipeAddr("local") }
func (c *namedPipeConn) RemoteAddr() net.Addr             { return pipeAddr("mihomo") }
func (c *namedPipeConn) SetDeadline(time.Time) error      { return nil }
func (c *namedPipeConn) SetReadDeadline(time.Time) error  { return nil }
func (c *namedPipeConn) SetWriteDeadline(time.Time) error { return nil }

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	comctl32 = syscall.NewLazyDLL("comctl32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")

	procRegisterClassExW     = user32.NewProc("RegisterClassExW")
	procCreateWindowExW      = user32.NewProc("CreateWindowExW")
	procDefWindowProcW       = user32.NewProc("DefWindowProcW")
	procShowWindow           = user32.NewProc("ShowWindow")
	procUpdateWindow         = user32.NewProc("UpdateWindow")
	procSetWindowPos          = user32.NewProc("SetWindowPos")
	procRtlMoveMemory          = kernel32.NewProc("RtlMoveMemory")
	procGetMessageW          = user32.NewProc("GetMessageW")
	procTranslateMessage     = user32.NewProc("TranslateMessage")
	procDispatchMessageW     = user32.NewProc("DispatchMessageW")
	procPostQuitMessage      = user32.NewProc("PostQuitMessage")
	procDestroyWindow        = user32.NewProc("DestroyWindow")
	procSendMessageW         = user32.NewProc("SendMessageW")
	procPostMessageW         = user32.NewProc("PostMessageW")
	procSetWindowTextW       = user32.NewProc("SetWindowTextW")
	procGetWindowTextW       = user32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
	procSetWindowLongPtrW    = user32.NewProc("SetWindowLongPtrW")
	procCallWindowProcW      = user32.NewProc("CallWindowProcW")
	procEnableWindow         = user32.NewProc("EnableWindow")
	procMessageBoxW          = user32.NewProc("MessageBoxW")
	procLoadCursorW          = user32.NewProc("LoadCursorW")
	procLoadIconW            = user32.NewProc("LoadIconW")
	procRegisterHotKey       = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey     = user32.NewProc("UnregisterHotKey")
	procSetForegroundWindow  = user32.NewProc("SetForegroundWindow")
	procFindWindowW          = user32.NewProc("FindWindowW")
	procGetForegroundWindow  = user32.NewProc("GetForegroundWindow")
	procGetClientRect        = user32.NewProc("GetClientRect")
	procIsWindowVisible      = user32.NewProc("IsWindowVisible")
	procCreatePopupMenu      = user32.NewProc("CreatePopupMenu")
	procAppendMenuW          = user32.NewProc("AppendMenuW")
	procTrackPopupMenu       = user32.NewProc("TrackPopupMenu")
	procDestroyMenu          = user32.NewProc("DestroyMenu")
	procGetCursorPos         = user32.NewProc("GetCursorPos")
	procGetModuleHandleW     = kernel32.NewProc("GetModuleHandleW")
	procGetStockObject       = gdi32.NewProc("GetStockObject")
	procInitCommonControlsEx = comctl32.NewProc("InitCommonControlsEx")
	procShellNotifyIconW     = shell32.NewProc("Shell_NotifyIconW")
)

var (
	hwndMain                                                                                                  syscall.Handle
	hController, hSecret, hConnect, hGroup, hSize, hScope, hPerRegion, hStart, hRetest, hStop, hList, hStatus syscall.Handle
	hAdvanced, hSettings, hRedetect, hConnLabel, hSecretLabel, hDetected, hEstimate                           syscall.Handle
	hPerRegionLabel, hRegionSelect, hNameFilter, hScopeDetail                                                 syscall.Handle
	hwndRegionPicker, hwndSettings                                                                            syscall.Handle
	hSetCloseTray, hSetStartTray, hSetHotkey, hSetHotkeyEdit                                                  syscall.Handle
	guiFont                                                                                                   syscall.Handle
	oldListWndProc                                                                                            uintptr
	listWndProcCallback                                                                                       uintptr

	mu                 sync.Mutex
	loadedProxies      map[string]proxyInfo
	groupNames         []string
	rows               []resultRow
	running            bool
	cancelRun          context.CancelFunc
	lastError          string
	closePending       bool
	advancedVisible    bool
	activeEndpoint     clashEndpoint
	detectedProxyPort  int
	detectedConfigPath string
	currentStatus      string
	runSeq             uint32
	activeRunID        uint32
	recoveryResult     string
	manualSwitching    bool
	singleRetest       bool
	switchResult       string
	switchTarget       string
	switchGroupName    string
	switchSucceeded    bool
	settings           appSettings
	hotkeyRegistered   bool
	appIconHandle      syscall.Handle
	trayAdded          bool
	selectedRegions    = map[string]bool{}
	regionCheckHandles = map[int]syscall.Handle{}
	regionCheckNames   = map[int]string{}
	lastScopeIndex     = -1
)

func u16(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }

func loword(v uintptr) uint16 { return uint16(v & 0xffff) }
func hiword(v uintptr) uint16 { return uint16((v >> 16) & 0xffff) }

func defaultSettings() appSettings {
	return appSettings{CloseToTray: true, StartMinimized: false, HotkeyEnabled: true, Hotkey: "Ctrl+Alt+B"}
}

func settingsFilePath() string {
	base := os.Getenv("APPDATA")
	if base == "" {
		if d, err := os.UserConfigDir(); err == nil {
			base = d
		}
	}
	if base == "" {
		return ""
	}
	return filepath.Join(base, "ClashBandwidthTest", "settings.json")
}

func loadSettings() appSettings {
	st := defaultSettings()
	path := settingsFilePath()
	if path == "" {
		return st
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return st
	}
	var saved appSettings
	if json.Unmarshal(b, &saved) != nil {
		return st
	}
	st.CloseToTray = saved.CloseToTray
	st.StartMinimized = saved.StartMinimized
	st.HotkeyEnabled = saved.HotkeyEnabled
	if strings.TrimSpace(saved.Hotkey) != "" {
		st.Hotkey = saved.Hotkey
	}
	return st
}

func saveSettings(st appSettings) error {
	path := settingsFilePath()
	if path == "" {
		return fmt.Errorf("无法确定设置目录")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0600)
}

func normalizeHotkeyText(s string) string {
	parts := strings.FieldsFunc(strings.TrimSpace(s), func(r rune) bool { return r == '+' || r == '＋' || r == ' ' })
	mods := map[string]bool{}
	key := ""
	for _, raw := range parts {
		t := strings.ToUpper(strings.TrimSpace(raw))
		switch t {
		case "CTRL", "CONTROL":
			mods["Ctrl"] = true
		case "ALT":
			mods["Alt"] = true
		case "SHIFT":
			mods["Shift"] = true
		case "WIN", "WINDOWS", "META":
			mods["Win"] = true
		default:
			if t != "" {
				key = t
			}
		}
	}
	var out []string
	for _, m := range []string{"Ctrl", "Alt", "Shift", "Win"} {
		if mods[m] {
			out = append(out, m)
		}
	}
	if key != "" {
		out = append(out, key)
	}
	return strings.Join(out, "+")
}

func parseHotkey(s string) (uint32, uint32, string, error) {
	norm := normalizeHotkeyText(s)
	if norm == "" {
		return 0, 0, "", fmt.Errorf("快捷键不能为空")
	}
	parts := strings.Split(norm, "+")
	var mods uint32
	key := ""
	for _, p := range parts {
		switch p {
		case "Ctrl":
			mods |= MOD_CONTROL
		case "Alt":
			mods |= MOD_ALT
		case "Shift":
			mods |= MOD_SHIFT
		case "Win":
			mods |= MOD_WIN
		default:
			key = strings.ToUpper(p)
		}
	}
	if mods == 0 {
		return 0, 0, "", fmt.Errorf("请至少使用 Ctrl、Alt、Shift 或 Win 中的一个修饰键")
	}
	var vk uint32
	if len(key) == 1 && ((key[0] >= 'A' && key[0] <= 'Z') || (key[0] >= '0' && key[0] <= '9')) {
		vk = uint32(key[0])
	} else if strings.HasPrefix(key, "F") {
		n, err := strconv.Atoi(strings.TrimPrefix(key, "F"))
		if err == nil && n >= 1 && n <= 12 {
			vk = uint32(0x70 + n - 1)
		}
	}
	if vk == 0 {
		return 0, 0, "", fmt.Errorf("主按键仅支持 A-Z、0-9 或 F1-F12")
	}
	return mods | MOD_NOREPEAT, vk, norm, nil
}

func unregisterHotkey() {
	if hwndMain != 0 && hotkeyRegistered {
		procUnregisterHotKey.Call(uintptr(hwndMain), HOTKEY_ID)
		hotkeyRegistered = false
	}
}

func registerHotkey(st appSettings) error {
	unregisterHotkey()
	if !st.HotkeyEnabled {
		return nil
	}
	mods, vk, norm, err := parseHotkey(st.Hotkey)
	if err != nil {
		return err
	}
	r, _, callErr := procRegisterHotKey.Call(uintptr(hwndMain), HOTKEY_ID, uintptr(mods), uintptr(vk))
	if r == 0 {
		if callErr != nil && callErr != syscall.Errno(0) {
			return fmt.Errorf("%s 注册失败：%v", norm, callErr)
		}
		return fmt.Errorf("%s 已被其他程序占用", norm)
	}
	hotkeyRegistered = true
	return nil
}

func fillUTF16(dst []uint16, text string) {
	v, _ := syscall.UTF16FromString(text)
	copy(dst, v)
}

func trayData() NOTIFYICONDATA {
	var nid NOTIFYICONDATA
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.HWnd = hwndMain
	nid.UID = TRAY_ICON_ID
	nid.UFlags = NIF_MESSAGE | NIF_ICON | NIF_TIP
	nid.UCallbackMessage = WM_APP_TRAY
	nid.HIcon = appIconHandle
	fillUTF16(nid.SzTip[:], "ClashBandwidthTest 1.8")
	return nid
}

func addTrayIcon() {
	if trayAdded || hwndMain == 0 {
		return
	}
	nid := trayData()
	r, _, _ := procShellNotifyIconW.Call(NIM_ADD, uintptr(unsafe.Pointer(&nid)))
	if r != 0 {
		trayAdded = true
		nid.UTimeoutOrVersion = NOTIFYICON_VERSION_4
		procShellNotifyIconW.Call(NIM_SETVERSION, uintptr(unsafe.Pointer(&nid)))
	}
}

func removeTrayIcon() {
	if !trayAdded {
		return
	}
	nid := trayData()
	procShellNotifyIconW.Call(NIM_DELETE, uintptr(unsafe.Pointer(&nid)))
	trayAdded = false
}

func restoreMainWindow() {
	if hwndMain == 0 {
		return
	}
	procShowWindow.Call(uintptr(hwndMain), SW_RESTORE)
	procSetForegroundWindow.Call(uintptr(hwndMain))
}

func toggleMainWindow() {
	if hwndMain == 0 {
		return
	}
	vis, _, _ := procIsWindowVisible.Call(uintptr(hwndMain))
	fg, _, _ := procGetForegroundWindow.Call()
	if vis != 0 && syscall.Handle(fg) == hwndMain {
		procShowWindow.Call(uintptr(hwndMain), SW_HIDE)
		return
	}
	restoreMainWindow()
}

func requestExit() {
	mu.Lock()
	c := cancelRun
	isRun := running
	runID := activeRunID
	if isRun {
		closePending = true
	}
	mu.Unlock()
	if isRun && c != nil {
		c()
		setRunStatus(runID, "正在停止测速并恢复原节点，完成后自动退出 ...")
		return
	}
	procDestroyWindow.Call(uintptr(hwndMain))
}

func setText(h syscall.Handle, s string) {
	procSetWindowTextW.Call(uintptr(h), uintptr(unsafe.Pointer(u16(s))))
}

func getText(h syscall.Handle) string {
	n, _, _ := procGetWindowTextLengthW.Call(uintptr(h))
	buf := make([]uint16, n+1)
	procGetWindowTextW.Call(uintptr(h), uintptr(unsafe.Pointer(&buf[0])), n+1)
	return syscall.UTF16ToString(buf)
}

func msgBox(title, text string, flags uintptr) {
	procMessageBoxW.Call(uintptr(hwndMain), uintptr(unsafe.Pointer(u16(text))), uintptr(unsafe.Pointer(u16(title))), flags)
}

func msgBoxResult(title, text string, flags uintptr) uintptr {
	r, _, _ := procMessageBoxW.Call(uintptr(hwndMain), uintptr(unsafe.Pointer(u16(text))), uintptr(unsafe.Pointer(u16(title))), flags)
	return r
}

func send(h syscall.Handle, msg uint32, w, l uintptr) uintptr {
	r, _, _ := procSendMessageW.Call(uintptr(h), uintptr(msg), w, l)
	return r
}

func selectedListIndex() int {
	if hList == 0 {
		return -1
	}
	r := send(hList, LVM_GETNEXTITEM, ^uintptr(0), LVNI_SELECTED)
	if int32(r) < 0 {
		return -1
	}
	return int(r)
}

func selectedResultRow() (resultRow, int, bool) {
	idx := selectedListIndex()
	if idx < 0 {
		return resultRow{}, -1, false
	}
	mu.Lock()
	defer mu.Unlock()
	if idx >= len(rows) {
		return resultRow{}, -1, false
	}
	return rows[idx], idx, true
}

func setLocalGroupNow(group, node string) {
	mu.Lock()
	defer mu.Unlock()
	if p, ok := loadedProxies[group]; ok {
		p.Now = node
		loadedProxies[group] = p
	}
}

func createControlOn(parent syscall.Handle, exStyle uint32, class, text string, style uint32, x, y, w, h int32, id int32) syscall.Handle {
	hInst, _, _ := procGetModuleHandleW.Call(0)
	r, _, _ := procCreateWindowExW.Call(
		uintptr(exStyle), uintptr(unsafe.Pointer(u16(class))), uintptr(unsafe.Pointer(u16(text))), uintptr(style),
		uintptr(x), uintptr(y), uintptr(w), uintptr(h), uintptr(parent), uintptr(id), hInst, 0,
	)
	hwnd := syscall.Handle(r)
	if hwnd != 0 && guiFont != 0 {
		send(hwnd, WM_SETFONT, uintptr(guiFont), 1)
	}
	return hwnd
}

func createControl(exStyle uint32, class, text string, style uint32, x, y, w, h int32, id int32) syscall.Handle {
	return createControlOn(hwndMain, exStyle, class, text, style, x, y, w, h, id)
}

func addColumn(index int32, name string, width int32) {
	t, _ := syscall.UTF16PtrFromString(name)
	col := LVCOLUMN{Mask: LVCF_FMT | LVCF_WIDTH | LVCF_TEXT, Fmt: LVCFMT_LEFT, Cx: width, PszText: t}
	send(hList, LVM_INSERTCOLUMNW, uintptr(index), uintptr(unsafe.Pointer(&col)))
}


func resizeListColumns() {
	if hList == 0 {
		return
	}
	var rc RECT
	r, _, _ := procGetClientRect.Call(uintptr(hList), uintptr(unsafe.Pointer(&rc)))
	if r == 0 {
		return
	}
	w := rc.Right - rc.Left
	if w < 400 {
		return
	}
	// Keep all columns visible when moving between monitors with different DPI.
	ratios := []int32{22, 8, 8, 8, 9, 11, 10, 10, 8, 6}
	for i, p := range ratios {
		cw := w * p / 100
		send(hList, LVM_SETCOLUMNWIDTH, uintptr(i), uintptr(cw))
	}
}


// scheduleListRebuild rebuilds the ListView after Windows finishes DPI/layout transitions.
// The native ListView control can temporarily lose its visual cache during monitor changes.
func scheduleListRebuild() {
	go func() {
		time.Sleep(500 * time.Millisecond)
		refreshList()
	}()
}

func refreshList() {
	mu.Lock()
	copyRows := append([]resultRow(nil), rows...)
	mu.Unlock()
	send(hList, LVM_DELETEALLITEMS, 0, 0)
	for i, r := range copyRows {
		vals := []string{r.Name, r.Region, r.Line, r.Type, formatLatency(r.Latency), formatSpeed(r.Speed), formatSpeed(r.P90), formatSpeed(r.Peak), r.Route, r.Status}
		t0, _ := syscall.UTF16PtrFromString(vals[0])
		item := LVITEM{Mask: LVIF_TEXT, IItem: int32(i), ISubItem: 0, PszText: t0}
		send(hList, LVM_INSERTITEMW, 0, uintptr(unsafe.Pointer(&item)))
		for j := 1; j < len(vals); j++ {
			tj, _ := syscall.UTF16PtrFromString(vals[j])
			sub := LVITEM{ISubItem: int32(j), PszText: tj}
			send(hList, LVM_SETITEMTEXTW, uintptr(i), uintptr(unsafe.Pointer(&sub)))
		}
	}
}

func formatLatency(ms int) string {
	if ms <= 0 {
		return "-"
	}
	return fmt.Sprintf("%d ms", ms)
}
func formatSpeed(s float64) string {
	if s <= 0 {
		return "-"
	}
	return fmt.Sprintf("%.2f MB/s", s)
}

func formatDataAmountMB(mb int) string {
	if mb >= 1000 {
		return fmt.Sprintf("%.2f GB", float64(mb)/1000.0)
	}
	return fmt.Sprintf("%d MB", mb)
}

func populateGroups() {
	send(hGroup, CB_RESETCONTENT, 0, 0)
	mu.Lock()
	names := append([]string(nil), groupNames...)
	mu.Unlock()
	best := 0
	bestScore := -1
	for i, n := range names {
		send(hGroup, CB_ADDSTRING, 0, uintptr(unsafe.Pointer(u16(n))))
		ln := strings.ToLower(n)
		score := 0
		for _, k := range []string{"节点选择", "proxy", "代理", "select", "proxies"} {
			if strings.Contains(ln, strings.ToLower(k)) {
				score += 10
			}
		}
		if strings.Contains(ln, "global") {
			score += 3
		}
		if score > bestScore {
			bestScore, best = score, i
		}
	}
	if len(names) > 0 {
		send(hGroup, CB_SETCURSEL, uintptr(best), 0)
	}
}

func selectedGroup() string {
	idx := int(send(hGroup, CB_GETCURSEL, 0, 0))
	mu.Lock()
	defer mu.Unlock()
	if idx >= 0 && idx < len(groupNames) {
		return groupNames[idx]
	}
	return ""
}

type testProfile struct {
	Name        string
	WarmupMB    int
	Streams     int
	PerStreamMB int
	Duration    time.Duration
	Settle      time.Duration
}

func (p testProfile) MaxMBPerNode() int {
	return p.WarmupMB + p.Streams*p.PerStreamMB
}

func selectedProfile() testProfile {
	idx := int(send(hSize, CB_GETCURSEL, 0, 0))
	if idx == 1 {
		return testProfile{
			Name:        "标准测速",
			WarmupMB:    2,
			Streams:     4,
			PerStreamMB: 15,
			Duration:    5 * time.Second,
			Settle:      800 * time.Millisecond,
		}
	}
	return testProfile{
		Name:        "快速筛选",
		WarmupMB:    1,
		Streams:     2,
		PerStreamMB: 6,
		Duration:    2 * time.Second,
		Settle:      500 * time.Millisecond,
	}
}

type regionDef struct {
	Name    string
	Aliases []string
}

var regionDefs = []regionDef{
	{"香港", []string{"香港", "hong kong", "hongkong", "hk"}},
	{"日本", []string{"日本", "japan", "tokyo", "osaka", "jp"}},
	{"新加坡", []string{"新加坡", "singapore", "sg"}},
	{"台湾", []string{"台湾", "taiwan", "taipei", "tw"}},
	{"美国", []string{"美国", "united states", "usa", "los angeles", "seattle", "san jose", "new york", "us"}},
	{"韩国", []string{"韩国", "south korea", "korea", "seoul", "kr"}},
	{"英国", []string{"英国", "united kingdom", "london", "uk"}},
	{"德国", []string{"德国", "germany", "frankfurt", "de"}},
	{"法国", []string{"法国", "france", "paris", "fr"}},
	{"荷兰", []string{"荷兰", "netherlands", "amsterdam", "nl"}},
	{"加拿大", []string{"加拿大", "canada", "toronto", "vancouver", "ca"}},
	{"澳大利亚", []string{"澳大利亚", "australia", "sydney", "melbourne", "au"}},
	{"俄罗斯", []string{"俄罗斯", "russia", "moscow", "ru"}},
	{"印度", []string{"印度", "india", "mumbai", "in"}},
	{"土耳其", []string{"土耳其", "turkey", "istanbul", "tr"}},
}

func asciiWordTokens(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	})
}

func containsAlias(name, alias string) bool {
	ln := strings.ToLower(name)
	la := strings.ToLower(alias)
	asciiOnly := true
	for _, r := range la {
		if r > 127 {
			asciiOnly = false
			break
		}
	}
	if !asciiOnly || strings.Contains(la, " ") {
		return strings.Contains(ln, la)
	}
	for _, t := range asciiWordTokens(ln) {
		if t == la {
			return true
		}
	}
	return false
}

func detectRegion(name string) string {
	for _, d := range regionDefs {
		for _, a := range d.Aliases {
			if containsAlias(name, a) {
				return d.Name
			}
		}
	}
	return "其他"
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func detectLine(name, region string) string {
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return !(r >= 'A' && r <= 'Z') && !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9')
	})
	regionCodes := map[string]bool{"hk": true, "jp": true, "sg": true, "tw": true, "us": true, "usa": true, "kr": true, "uk": true, "de": true, "fr": true, "nl": true, "ca": true, "au": true, "ru": true, "in": true, "tr": true}
	for _, raw := range parts {
		t := strings.ToLower(strings.TrimSpace(raw))
		if t == "" || regionCodes[t] || isDigits(t) {
			continue
		}
		if len(t) == 1 && t[0] >= 'a' && t[0] <= 'z' {
			continue
		}
		if len(t) >= 2 && len(t) <= 10 {
			return strings.ToUpper(t)
		}
	}
	return "—"
}

func selectedNodeNames() []string {
	group := selectedGroup()
	if group == "" {
		return nil
	}
	mu.Lock()
	defer mu.Unlock()
	gp, ok := loadedProxies[group]
	if !ok {
		return nil
	}
	var out []string
	for _, name := range gp.All {
		if p, exists := loadedProxies[name]; exists && isLeafProxy(p) {
			out = append(out, name)
		}
	}
	return out
}

const (
	scopeSmart = iota
	scopeRegions
	scopeFilter
	scopeAll
)

func selectedScope() int {
	idx := int(send(hScope, CB_GETCURSEL, 0, 0))
	if idx < scopeSmart || idx > scopeAll {
		return scopeSmart
	}
	return idx
}

// 0 means "all nodes in each selected region".
func selectedPerRegion() int {
	idx := int(send(hPerRegion, CB_GETCURSEL, 0, 0))
	switch idx {
	case 1:
		return 1
	case 2:
		return 2
	case 3:
		return 3
	default:
		return 0
	}
}

func regionCandidateEstimate(nodes []string, perRegion int) (regions int, candidates int) {
	counts := map[string]int{}
	for _, n := range nodes {
		counts[detectRegion(n)]++
	}
	regions = len(counts)
	for _, c := range counts {
		if perRegion <= 0 || c <= perRegion {
			candidates += c
		} else {
			candidates += perRegion
		}
	}
	return
}

func availableRegionCounts() map[string]int {
	counts := map[string]int{}
	for _, n := range selectedNodeNames() {
		counts[detectRegion(n)]++
	}
	return counts
}

func orderedAvailableRegions() []string {
	counts := availableRegionCounts()
	var out []string
	for _, d := range regionDefs {
		if counts[d.Name] > 0 {
			out = append(out, d.Name)
		}
	}
	if counts["其他"] > 0 {
		out = append(out, "其他")
	}
	return out
}

func pruneSelectedRegions() {
	available := availableRegionCounts()
	for r := range selectedRegions {
		if available[r] == 0 {
			delete(selectedRegions, r)
		}
	}
}

func updateRegionButton() {
	if hRegionSelect == 0 {
		return
	}
	pruneSelectedRegions()
	var names []string
	for _, r := range orderedAvailableRegions() {
		if selectedRegions[r] {
			names = append(names, r)
		}
	}
	text := "选择地区…"
	if len(names) == 1 {
		text = names[0]
	} else if len(names) <= 3 && len(names) > 1 {
		text = strings.Join(names, "、")
	} else if len(names) > 3 {
		text = fmt.Sprintf("已选 %d 个地区", len(names))
	}
	setText(hRegionSelect, text)
}

func selectedRegionSetCopy() map[string]bool {
	out := map[string]bool{}
	for k, v := range selectedRegions {
		if v {
			out[k] = true
		}
	}
	return out
}

func splitFilterTerms(q string) []string {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil
	}
	parts := strings.FieldsFunc(q, func(r rune) bool {
		switch r {
		case ',', '，', ';', '；', '|':
			return true
		}
		return false
	})
	var out []string
	for _, p := range parts {
		p = strings.ToLower(strings.TrimSpace(p))
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func filterNodesByName(nodes []string, q string) []string {
	terms := splitFilterTerms(q)
	if len(terms) == 0 {
		return nil
	}
	var out []string
	for _, n := range nodes {
		ln := strings.ToLower(n)
		for _, t := range terms {
			if strings.Contains(ln, t) {
				out = append(out, n)
				break
			}
		}
	}
	return out
}

func targetNodesForScope(all []string) ([]string, string) {
	switch selectedScope() {
	case scopeRegions:
		regions := selectedRegionSetCopy()
		if len(regions) == 0 {
			return nil, "请先选择至少一个地区。"
		}
		var out []string
		for _, n := range all {
			if regions[detectRegion(n)] {
				out = append(out, n)
			}
		}
		if len(out) == 0 {
			return nil, "所选地区在当前策略组中没有可测速节点。"
		}
		return out, ""
	case scopeFilter:
		q := strings.TrimSpace(getText(hNameFilter))
		if q == "" {
			return nil, "请输入节点名称筛选词。可用逗号或 | 表示多个条件。"
		}
		out := filterNodesByName(all, q)
		if len(out) == 0 {
			return nil, "没有节点匹配当前名称筛选。"
		}
		return out, ""
	default:
		return append([]string(nil), all...), ""
	}
}

func scopeDescriptionForRun(nodes []string, perRegion int) string {
	switch selectedScope() {
	case scopeSmart:
		regions, estimated := regionCandidateEstimate(nodes, perRegion)
		if perRegion <= 0 {
			return fmt.Sprintf("地区智能：%d 个地区，精测全部 %d 个节点", regions, len(nodes))
		}
		return fmt.Sprintf("地区智能抽样：%d 个地区，预计精测 %d 个节点", regions, estimated)
	case scopeRegions:
		regions, estimated := regionCandidateEstimate(nodes, perRegion)
		if perRegion <= 0 {
			return fmt.Sprintf("指定地区：%d 个地区，共 %d 个节点全部精测", regions, len(nodes))
		}
		return fmt.Sprintf("指定地区：%d 个地区，每地区最多 %d 个，预计精测 %d 个", regions, perRegion, estimated)
	case scopeFilter:
		return fmt.Sprintf("名称筛选：匹配 %d 个节点", len(nodes))
	default:
		return fmt.Sprintf("全部节点：%d 个", len(nodes))
	}
}

func updateScopeControls() {
	if hScope == 0 {
		return
	}
	scope := selectedScope()
	showWindow(hRegionSelect, scope == scopeRegions)
	showWindow(hNameFilter, scope == scopeFilter)
	showWindow(hScopeDetail, scope == scopeSmart || scope == scopeAll)
	if scope == scopeSmart {
		setText(hScopeDetail, "自动识别全部地区")
		setText(hPerRegionLabel, "每地区")
		procEnableWindow.Call(uintptr(hPerRegion), 1)
		if lastScopeIndex != scopeSmart {
			send(hPerRegion, CB_SETCURSEL, 2, 0) // 2 个
		}
	} else if scope == scopeRegions {
		setText(hPerRegionLabel, "地区内")
		procEnableWindow.Call(uintptr(hPerRegion), 1)
		if lastScopeIndex != scopeRegions {
			send(hPerRegion, CB_SETCURSEL, 0, 0) // 默认全部
		}
		updateRegionButton()
	} else {
		setText(hPerRegionLabel, "地区内")
		procEnableWindow.Call(uintptr(hPerRegion), 0)
		if scope == scopeAll {
			setText(hScopeDetail, "全部可测速节点")
		}
	}
	lastScopeIndex = scope
}

func updateEstimate() {
	if hEstimate == 0 {
		return
	}
	all := selectedNodeNames()
	profile := selectedProfile()
	if len(all) == 0 {
		setText(hEstimate, "请选择策略组。")
		return
	}
	nodes, hint := targetNodesForScope(all)
	if hint != "" {
		setText(hEstimate, hint)
		return
	}
	perRegion := selectedPerRegion()
	scope := selectedScope()
	if scope == scopeSmart || scope == scopeRegions {
		regions, candidates := regionCandidateEstimate(nodes, perRegion)
		total := candidates * profile.MaxMBPerNode()
		mode := "地区智能"
		if scope == scopeRegions {
			mode = "指定地区"
		}
		if perRegion <= 0 {
			setText(hEstimate, fmt.Sprintf("%s：%d 个节点 / %d 个地区 → 全部精测；最多约 %s（另有少量延迟/验证流量）",
				mode, len(nodes), regions, formatDataAmountMB(total)))
		} else {
			setText(hEstimate, fmt.Sprintf("%s：%d 个节点 / %d 个地区 → 预计精测 %d 个；最多约 %s（另有少量延迟/验证流量）",
				mode, len(nodes), regions, candidates, formatDataAmountMB(total)))
		}
		return
	}
	total := len(nodes) * profile.MaxMBPerNode()
	maxSeconds := int(profile.Duration.Seconds()) * len(nodes)
	prefix := "全部节点"
	if scope == scopeFilter {
		prefix = "名称筛选"
	}
	setText(hEstimate, fmt.Sprintf("%s：%d 个 · 每节点预热 %d MB + %d 路并发 · 最多约 %s · 纯测速时间上限约 %s",
		prefix, len(nodes), profile.WarmupMB, profile.Streams, formatDataAmountMB(total), formatDuration(maxSeconds)))
}

func formatDuration(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%d 秒", seconds)
	}
	m := seconds / 60
	s := seconds % 60
	if s == 0 {
		return fmt.Sprintf("%d 分钟", m)
	}
	return fmt.Sprintf("%d 分 %d 秒", m, s)
}

func setRunning(v bool) {
	mu.Lock()
	running = v
	mu.Unlock()
	if v {
		procEnableWindow.Call(uintptr(hConnect), 0)
		procEnableWindow.Call(uintptr(hRedetect), 0)
		procEnableWindow.Call(uintptr(hAdvanced), 0)
		procEnableWindow.Call(uintptr(hStart), 0)
		procEnableWindow.Call(uintptr(hRetest), 0)
		procEnableWindow.Call(uintptr(hGroup), 0)
		procEnableWindow.Call(uintptr(hSize), 0)
		procEnableWindow.Call(uintptr(hScope), 0)
		procEnableWindow.Call(uintptr(hPerRegion), 0)
		procEnableWindow.Call(uintptr(hRegionSelect), 0)
		procEnableWindow.Call(uintptr(hNameFilter), 0)
		procEnableWindow.Call(uintptr(hStop), 1)
	} else {
		procEnableWindow.Call(uintptr(hConnect), 1)
		procEnableWindow.Call(uintptr(hRedetect), 1)
		procEnableWindow.Call(uintptr(hAdvanced), 1)
		procEnableWindow.Call(uintptr(hStart), 1)
		mu.Lock()
		hasRows := len(rows) > 0
		mu.Unlock()
		if hasRows {
			procEnableWindow.Call(uintptr(hRetest), 1)
		}
		procEnableWindow.Call(uintptr(hGroup), 1)
		procEnableWindow.Call(uintptr(hSize), 1)
		procEnableWindow.Call(uintptr(hScope), 1)
		procEnableWindow.Call(uintptr(hRegionSelect), 1)
		procEnableWindow.Call(uintptr(hNameFilter), 1)
		procEnableWindow.Call(uintptr(hStop), 0)
		updateScopeControls()
	}
}

func showWindow(h syscall.Handle, show bool) {
	cmd := uintptr(SW_HIDE)
	if show {
		cmd = uintptr(SW_SHOW)
	}
	procShowWindow.Call(uintptr(h), cmd)
}

func setAdvancedVisible(v bool) {
	advancedVisible = v
	for _, h := range []syscall.Handle{hConnLabel, hController, hSecretLabel, hSecret, hConnect, hRedetect} {
		showWindow(h, v)
	}
	if v {
		setText(hAdvanced, "收起高级")
	} else {
		setText(hAdvanced, "高级")
	}
}

func cleanYAMLScalar(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, " #"); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		if u, err := strconv.Unquote(s); err == nil {
			return strings.TrimSpace(u)
		}
		s = s[1 : len(s)-1]
	} else if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		s = s[1 : len(s)-1]
	}
	return strings.TrimSpace(s)
}

func parseRuntimeConfig(path string) (detectedConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return detectedConfig{}, err
	}
	defer f.Close()
	cfg := detectedConfig{Path: path}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := cleanYAMLScalar(parts[1])
		switch key {
		case "external-controller":
			cfg.Controller = val
		case "external-controller-pipe":
			cfg.Pipe = val
		case "secret":
			cfg.Secret = val
		case "mixed-port":
			if n, e := strconv.Atoi(val); e == nil {
				cfg.MixedPort = n
			}
		}
	}
	if err := sc.Err(); err != nil {
		return detectedConfig{}, err
	}
	return cfg, nil
}

func candidateConfigPaths() []string {
	var out []string
	seen := map[string]bool{}
	add := func(p string) {
		if p == "" || seen[p] {
			return
		}
		seen[p] = true
		out = append(out, p)
	}
	appData := os.Getenv("APPDATA")
	if appData == "" {
		if d, err := os.UserConfigDir(); err == nil {
			appData = d
		}
	}
	if appData == "" {
		return out
	}
	for _, d := range []string{
		"io.github.clash-verge-rev.clash-verge-rev",
		"clash-verge-rev",
		"Clash Verge Rev",
	} {
		add(filepath.Join(appData, d, "config.yaml"))
	}
	entries, _ := os.ReadDir(appData)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		n := strings.ToLower(e.Name())
		if strings.Contains(n, "clash") && strings.Contains(n, "verge") {
			add(filepath.Join(appData, e.Name(), "config.yaml"))
		}
	}
	return out
}

func normalizeController(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	if strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://") {
		return v
	}
	if strings.HasPrefix(v, "0.0.0.0:") {
		v = "127.0.0.1:" + strings.TrimPrefix(v, "0.0.0.0:")
	}
	if strings.HasPrefix(v, "[::]:") {
		v = "127.0.0.1:" + strings.TrimPrefix(v, "[::]:")
	}
	if strings.HasPrefix(v, ":") {
		v = "127.0.0.1" + v
	}
	return "http://" + v
}

func dialNamedPipe(ctx context.Context, pipe string) (net.Conn, error) {
	p, err := syscall.UTF16PtrFromString(pipe)
	if err != nil {
		return nil, err
	}
	const (
		genericRead    = 0x80000000
		genericWrite   = 0x40000000
		openExisting   = 3
		fileAttrNormal = 0x80
	)
	var last error
	for i := 0; i < 20; i++ {
		h, e := syscall.CreateFile(p, genericRead|genericWrite, 0, nil, openExisting, fileAttrNormal, 0)
		if e == nil {
			return &namedPipeConn{h: h}, nil
		}
		last = e
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(75 * time.Millisecond):
		}
	}
	return nil, last
}

func endpointClient(ep clashEndpoint, timeout time.Duration) *http.Client {
	if ep.Pipe != "" {
		tr := &http.Transport{
			DisableKeepAlives: true,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialNamedPipe(ctx, ep.Pipe)
			},
		}
		return &http.Client{Transport: tr, Timeout: timeout}
	}
	return &http.Client{Timeout: timeout}
}

func endpointRequest(ctx context.Context, ep clashEndpoint, method, path string, body []byte) ([]byte, int, error) {
	base := ep.Base
	if ep.Pipe != "" {
		base = "http://mihomo.local"
	}
	if base == "" {
		return nil, 0, fmt.Errorf("没有可用的 Clash 控制端点")
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(base, "/")+path, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	if ep.Secret != "" {
		req.Header.Set("Authorization", "Bearer "+ep.Secret)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := endpointClient(ep, 10*time.Second).Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return b, resp.StatusCode, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	return b, resp.StatusCode, nil
}

func autoDetectEndpoint() (clashEndpoint, int, string, error) {
	var problems []string
	for _, path := range candidateConfigPaths() {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		cfg, err := parseRuntimeConfig(path)
		if err != nil {
			problems = append(problems, filepath.Base(filepath.Dir(path))+": "+err.Error())
			continue
		}
		// Prefer the private Windows named pipe used by Clash Verge Rev itself.
		if cfg.Pipe != "" {
			ep := clashEndpoint{Pipe: cfg.Pipe, Secret: cfg.Secret, Mode: "Windows Named Pipe"}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			_, _, e := endpointRequest(ctx, ep, http.MethodGet, "/proxies", nil)
			cancel()
			if e == nil {
				return ep, cfg.MixedPort, path, nil
			}
			problems = append(problems, "Named Pipe: "+e.Error())
		}
		if cfg.Controller != "" {
			ep := clashEndpoint{Base: normalizeController(cfg.Controller), Secret: cfg.Secret, Mode: "HTTP Controller"}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			_, _, e := endpointRequest(ctx, ep, http.MethodGet, "/proxies", nil)
			cancel()
			if e == nil {
				return ep, cfg.MixedPort, path, nil
			}
			problems = append(problems, "HTTP Controller: "+e.Error())
		}
	}
	// Last-resort local-only probe for users who already enabled the common controller port.
	for _, base := range []string{"http://127.0.0.1:9090", "http://127.0.0.1:9097"} {
		ep := clashEndpoint{Base: base, Mode: "HTTP Controller"}
		ctx, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
		_, _, e := endpointRequest(ctx, ep, http.MethodGet, "/proxies", nil)
		cancel()
		if e == nil {
			return ep, 0, "", nil
		}
	}
	if len(problems) > 0 {
		return clashEndpoint{}, 0, "", fmt.Errorf("自动检测失败：%s", strings.Join(problems, "；"))
	}
	return clashEndpoint{}, 0, "", fmt.Errorf("自动检测失败：未找到 Clash Verge Rev 的运行配置")
}

func controllerRequest(ctx context.Context, method, base, secret, path string, body []byte) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(base, "/")+path, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	if secret != "" {
		req.Header.Set("Authorization", "Bearer "+secret)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return b, resp.StatusCode, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	return b, resp.StatusCode, nil
}

func loadClash(ep clashEndpoint) (map[string]proxyInfo, []string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	b, _, err := endpointRequest(ctx, ep, http.MethodGet, "/proxies", nil)
	if err != nil {
		return nil, nil, err
	}
	var pr proxiesResponse
	if err := json.Unmarshal(b, &pr); err != nil {
		return nil, nil, err
	}
	var groups []string
	for name, p := range pr.Proxies {
		if strings.EqualFold(p.Type, "Selector") && len(p.All) > 0 {
			groups = append(groups, name)
		}
	}
	sort.Strings(groups)
	if len(groups) == 0 {
		return pr.Proxies, nil, fmt.Errorf("没有找到 Selector 策略组")
	}
	return pr.Proxies, groups, nil
}

func connectAsync() {
	base := strings.TrimSpace(getText(hController))
	secret := strings.TrimSpace(getText(hSecret))
	if base == "" {
		base = "http://127.0.0.1:9090"
		setText(hController, base)
	}
	ep := clashEndpoint{Base: normalizeController(base), Secret: secret, Mode: "HTTP Controller"}
	setText(hStatus, "正在手动连接 Clash/Mihomo ...")
	procEnableWindow.Call(uintptr(hConnect), 0)
	go func() {
		p, g, err := loadClash(ep)
		mu.Lock()
		if err != nil {
			lastError = err.Error()
		} else {
			loadedProxies, groupNames, lastError = p, g, ""
			activeEndpoint = ep
			detectedProxyPort = 0
			detectedConfigPath = ""
		}
		mu.Unlock()
		procPostMessageW.Call(uintptr(hwndMain), WM_APP_GROUPS, 0, 0)
	}()
}

func autoDetectAsync() {
	setText(hDetected, "● 正在自动检测 Clash Verge Rev ...")
	setText(hStatus, "正在读取 Clash Verge Rev 运行配置并连接 Mihomo ...")
	procEnableWindow.Call(uintptr(hRedetect), 0)
	go func() {
		ep, pport, cfgPath, err := autoDetectEndpoint()
		var p map[string]proxyInfo
		var g []string
		if err == nil {
			p, g, err = loadClash(ep)
		}
		mu.Lock()
		if err != nil {
			lastError = err.Error()
		} else {
			loadedProxies, groupNames, lastError = p, g, ""
			activeEndpoint = ep
			detectedProxyPort = pport
			detectedConfigPath = cfgPath
		}
		mu.Unlock()
		procPostMessageW.Call(uintptr(hwndMain), WM_APP_DETECT, 0, 0)
	}()
}

func isLeafProxy(p proxyInfo) bool {
	t := strings.ToLower(p.Type)
	switch t {
	case "selector", "urltest", "fallback", "loadbalance", "compatible", "pass", "reject", "direct":
		return false
	}
	return true
}

func switchGroup(ctx context.Context, ep clashEndpoint, group, node string) error {
	payload, _ := json.Marshal(map[string]string{"name": node})
	_, _, err := endpointRequest(ctx, ep, http.MethodPut, "/proxies/"+url.PathEscape(group), payload)
	return err
}

func getProxyPort(ctx context.Context, ep clashEndpoint, fallback int) (int, error) {
	b, _, err := endpointRequest(ctx, ep, http.MethodGet, "/configs", nil)
	if err != nil {
		return 0, err
	}
	var c configResponse
	if err := json.Unmarshal(b, &c); err != nil {
		return 0, err
	}
	if c.MixedPort > 0 {
		return c.MixedPort, nil
	}
	if c.Port > 0 {
		return c.Port, nil
	}
	if fallback > 0 {
		return fallback, nil
	}
	return 0, fmt.Errorf("Clash 未启用 mixed-port 或 HTTP port")
}

func measureLatency(ctx context.Context, ep clashEndpoint, node string) int {
	q := "/proxies/" + url.PathEscape(node) + "/delay?timeout=5000&url=" + url.QueryEscape("https://www.gstatic.com/generate_204")
	b, _, err := endpointRequest(ctx, ep, http.MethodGet, q, nil)
	if err != nil {
		return 0
	}
	var x struct {
		Delay int `json:"delay"`
	}
	if json.Unmarshal(b, &x) != nil {
		return 0
	}
	return x.Delay
}

func getConnections(ctx context.Context, ep clashEndpoint) (connectionSnapshot, error) {
	var snap connectionSnapshot
	b, _, err := endpointRequest(ctx, ep, http.MethodGet, "/connections", nil)
	if err != nil {
		return snap, err
	}
	if err := json.Unmarshal(b, &snap); err != nil {
		return snap, err
	}
	return snap, nil
}

func existingHostConnectionIDs(ctx context.Context, ep clashEndpoint, host string) map[string]bool {
	out := map[string]bool{}
	snap, err := getConnections(ctx, ep)
	if err != nil {
		return out
	}
	for _, c := range snap.Connections {
		if strings.EqualFold(strings.TrimSuffix(c.Metadata.Host, "."), host) {
			out[c.ID] = true
		}
	}
	return out
}

func chainContains(chains []string, node string) bool {
	for _, c := range chains {
		if c == node {
			return true
		}
	}
	return false
}

func verifyNewHostConnection(ctx context.Context, ep clashEndpoint, host, node string, before map[string]bool) (found bool, ok bool, desc string) {
	for attempt := 0; attempt < 8; attempt++ {
		snap, err := getConnections(ctx, ep)
		if err == nil {
			for _, c := range snap.Connections {
				if before[c.ID] {
					continue
				}
				if !strings.EqualFold(strings.TrimSuffix(c.Metadata.Host, "."), host) {
					continue
				}
				chain := strings.Join(c.Chains, " → ")
				if chain == "" {
					chain = c.Rule
				}
				if chainContains(c.Chains, node) {
					return true, true, "✓ " + node
				}
				if chain == "" {
					chain = "未知链路"
				}
				return true, false, "⚠ " + chain
			}
		}
		select {
		case <-ctx.Done():
			return false, false, "? 未验证"
		case <-time.After(75 * time.Millisecond):
		}
	}
	return false, false, "? 未捕获"
}

type speedResult struct {
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

func newSpeedClient(proxyPort int) (*http.Client, *http.Transport) {
	proxyURL, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", proxyPort))
	tr := &http.Transport{
		Proxy:                 http.ProxyURL(proxyURL),
		DisableKeepAlives:     true,
		TLSHandshakeTimeout:   8 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ForceAttemptHTTP2:     false,
	}
	return &http.Client{Transport: tr}, tr
}

func readAtMost(ctx context.Context, body io.Reader, want int64) (int64, error) {
	buf := make([]byte, 128*1024)
	var total int64
	for total < want {
		select {
		case <-ctx.Done():
			return total, ctx.Err()
		default:
		}
		next := int64(len(buf))
		if remain := want - total; remain < next {
			next = remain
		}
		n, err := body.Read(buf[:int(next)])
		if n > 0 {
			total += int64(n)
		}
		if err != nil {
			if err == io.EOF {
				return total, nil
			}
			return total, err
		}
	}
	return total, nil
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

func measureSpeed(ctx context.Context, ep clashEndpoint, node string, proxyPort int, profile testProfile) (speedResult, error) {
	var out speedResult
	host := "speed.cloudflare.com"
	dctx, cancel := context.WithTimeout(ctx, profile.Duration+20*time.Second)
	defer cancel()

	// Route probe + warm-up. The warm-up bytes are deliberately excluded from
	// the reported throughput so TCP/TLS setup and slow-start distort the result less.
	warmBytes := int64(profile.WarmupMB) * 1000 * 1000
	before := existingHostConnectionIDs(dctx, ep, host)
	warmClient, warmTr := newSpeedClient(proxyPort)
	defer warmTr.CloseIdleConnections()
	warmTarget := "https://" + host + "/__down?bytes=" + strconv.FormatInt(warmBytes, 10)
	req, err := http.NewRequestWithContext(dctx, http.MethodGet, warmTarget, nil)
	if err != nil {
		return out, err
	}
	req.Header.Set("User-Agent", "ClashBandwidthTest/1.6")
	req.Header.Set("Cache-Control", "no-cache")
	resp, err := warmClient.Do(req)
	if err != nil {
		return out, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return out, fmt.Errorf("测速源 HTTP %d", resp.StatusCode)
	}
	found, routeOK, routeDesc := verifyNewHostConnection(dctx, ep, host, node, before)
	out.RouteFound, out.RouteOK, out.Route = found, routeOK, routeDesc
	if found && !routeOK {
		resp.Body.Close()
		return out, fmt.Errorf("测速流量没有经过当前节点：%s", routeDesc)
	}
	if warmBytes > 0 {
		got, rerr := readAtMost(dctx, resp.Body, warmBytes)
		resp.Body.Close()
		if rerr != nil && dctx.Err() == nil {
			return out, rerr
		}
		if got < 100000 {
			return out, fmt.Errorf("预热下载数据过少")
		}
	} else {
		resp.Body.Close()
	}

	// Open independent HTTP streams first, then release them together. This makes
	// the timed section much less sensitive to connection-establishment latency.
	// Stream requests use a cancellable context so the time limit interrupts a
	// blocked Body.Read immediately rather than waiting on the remote server.
	benchCtx, benchCancel := context.WithCancel(dctx)
	defer benchCancel()
	type openResult struct {
		s   openedStream
		err error
	}
	openCh := make(chan openResult, profile.Streams)
	streamBytes := int64(profile.PerStreamMB) * 1000 * 1000
	for i := 0; i < profile.Streams; i++ {
		go func() {
			client, tr := newSpeedClient(proxyPort)
			target := "https://" + host + "/__down?bytes=" + strconv.FormatInt(streamBytes, 10)
			rq, e := http.NewRequestWithContext(benchCtx, http.MethodGet, target, nil)
			if e != nil {
				tr.CloseIdleConnections()
				openCh <- openResult{err: e}
				return
			}
			rq.Header.Set("User-Agent", "ClashBandwidthTest/1.6")
			rq.Header.Set("Cache-Control", "no-cache")
			rp, e := client.Do(rq)
			if e != nil {
				tr.CloseIdleConnections()
				openCh <- openResult{err: e}
				return
			}
			if rp.StatusCode < 200 || rp.StatusCode >= 300 {
				rp.Body.Close()
				tr.CloseIdleConnections()
				openCh <- openResult{err: fmt.Errorf("测速源 HTTP %d", rp.StatusCode)}
				return
			}
			openCh <- openResult{s: openedStream{body: rp.Body, tr: tr}}
		}()
	}
	streams := make([]openedStream, 0, profile.Streams)
	var firstOpenErr error
	for i := 0; i < profile.Streams; i++ {
		r := <-openCh
		if r.err != nil {
			if firstOpenErr == nil {
				firstOpenErr = r.err
			}
			continue
		}
		streams = append(streams, r.s)
	}
	if len(streams) == 0 {
		if firstOpenErr != nil {
			return out, firstOpenErr
		}
		return out, fmt.Errorf("无法建立测速连接")
	}

	var total int64
	var readers sync.WaitGroup
	readers.Add(len(streams))
	for _, st := range streams {
		st := st
		go func() {
			defer readers.Done()
			defer st.body.Close()
			defer st.tr.CloseIdleConnections()
			buf := make([]byte, 128*1024)
			for {
				select {
				case <-benchCtx.Done():
					return
				default:
				}
				n, er := st.body.Read(buf)
				if n > 0 {
					atomic.AddInt64(&total, int64(n))
				}
				if er != nil {
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
	if bytesDone < 100000 {
		return out, fmt.Errorf("下载数据过少")
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

type recoveryState struct {
	Group     string `json:"group"`
	Original  string `json:"original"`
	StartedAt string `json:"started_at"`
}

func recoveryFilePath() string {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		if d, err := os.UserConfigDir(); err == nil {
			base = d
		}
	}
	if base == "" {
		return ""
	}
	return filepath.Join(base, "ClashBandwidthTest", "recovery.json")
}

func saveRecoveryState(group, original string) error {
	path := recoveryFilePath()
	if path == "" {
		return fmt.Errorf("无法确定本机配置目录")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	st := recoveryState{Group: group, Original: original, StartedAt: time.Now().Format(time.RFC3339)}
	b, _ := json.MarshalIndent(st, "", "  ")
	return os.WriteFile(path, b, 0600)
}

func loadRecoveryState() (recoveryState, error) {
	path := recoveryFilePath()
	if path == "" {
		return recoveryState{}, os.ErrNotExist
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return recoveryState{}, err
	}
	var st recoveryState
	if err := json.Unmarshal(b, &st); err != nil {
		return recoveryState{}, err
	}
	if st.StartedAt != "" {
		if t, err := time.Parse(time.RFC3339, st.StartedAt); err == nil && time.Since(t) > 7*24*time.Hour {
			_ = os.Remove(path)
			return recoveryState{}, os.ErrNotExist
		}
	}
	return st, nil
}

func clearRecoveryState() {
	if path := recoveryFilePath(); path != "" {
		_ = os.Remove(path)
	}
}

func maybeOfferRecovery() {
	st, err := loadRecoveryState()
	if err != nil || st.Group == "" || st.Original == "" {
		return
	}
	mu.Lock()
	ep := activeEndpoint
	gp, groupOK := loadedProxies[st.Group]
	_, nodeOK := loadedProxies[st.Original]
	mu.Unlock()
	if !groupOK || !nodeOK {
		clearRecoveryState()
		return
	}
	if gp.Now == st.Original {
		clearRecoveryState()
		return
	}
	text := fmt.Sprintf("检测到上一次测速可能异常结束。\r\n\r\n策略组：%s\r\n测速前节点：%s\r\n当前节点：%s\r\n\r\n是否现在恢复测速前的节点？", st.Group, st.Original, gp.Now)
	if msgBoxResult(appTitle, text, MB_YESNO|MB_ICONWARNING) != IDYES {
		clearRecoveryState()
		return
	}
	setText(hStatus, "正在恢复上一次测速前的节点 ...")
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		err := switchGroup(ctx, ep, st.Group, st.Original)
		cancel()
		mu.Lock()
		if err != nil {
			recoveryResult = "恢复失败：" + err.Error()
		} else {
			recoveryResult = "已恢复上一次测速前的节点：" + st.Original
			clearRecoveryState()
		}
		mu.Unlock()
		procPostMessageW.Call(uintptr(hwndMain), WM_APP_RECOVER, 0, 0)
	}()
}

type regionalPick struct {
	Name    string
	Region  string
	Line    string
	Latency int
	Order   int
}

func betterLatency(a, b regionalPick) bool {
	if a.Latency <= 0 && b.Latency > 0 {
		return false
	}
	if a.Latency > 0 && b.Latency <= 0 {
		return true
	}
	if a.Latency == b.Latency {
		return a.Order < b.Order
	}
	return a.Latency < b.Latency
}

// selectRegionalCandidates keeps geographic diversity first, then line diversity
// inside each region. The returned order is round-robin by region so the user gets
// useful cross-region results early instead of waiting for one region to finish.
func selectRegionalCandidates(nodes []string, latency map[string]int, perRegion int) []string {
	groups := map[string][]regionalPick{}
	var regionOrder []string
	seenRegion := map[string]bool{}
	for i, n := range nodes {
		r := detectRegion(n)
		l := detectLine(n, r)
		if !seenRegion[r] {
			seenRegion[r] = true
			regionOrder = append(regionOrder, r)
		}
		groups[r] = append(groups[r], regionalPick{Name: n, Region: r, Line: l, Latency: latency[n], Order: i})
	}

	picksByRegion := map[string][]regionalPick{}
	for _, r := range regionOrder {
		arr := append([]regionalPick(nil), groups[r]...)
		sort.SliceStable(arr, func(i, j int) bool { return betterLatency(arr[i], arr[j]) })
		k := perRegion
		if k <= 0 || k > len(arr) {
			k = len(arr)
		}
		chosen := make([]regionalPick, 0, k)
		chosenName := map[string]bool{}

		// First take the best node from each line family.
		bestLine := map[string]regionalPick{}
		for _, x := range arr {
			line := x.Line
			if line == "" || line == "—" {
				line = "默认"
			}
			cur, ok := bestLine[line]
			if !ok || betterLatency(x, cur) {
				bestLine[line] = x
			}
		}
		var lineBest []regionalPick
		for _, x := range bestLine {
			lineBest = append(lineBest, x)
		}
		sort.SliceStable(lineBest, func(i, j int) bool { return betterLatency(lineBest[i], lineBest[j]) })
		for _, x := range lineBest {
			if len(chosen) >= k {
				break
			}
			chosen = append(chosen, x)
			chosenName[x.Name] = true
		}
		// Then fill any remaining slots by latency.
		for _, x := range arr {
			if len(chosen) >= k {
				break
			}
			if chosenName[x.Name] {
				continue
			}
			chosen = append(chosen, x)
			chosenName[x.Name] = true
		}
		picksByRegion[r] = chosen
	}

	var out []string
	for round := 0; ; round++ {
		added := false
		for _, r := range regionOrder {
			if round < len(picksByRegion[r]) {
				out = append(out, picksByRegion[r][round].Name)
				added = true
			}
		}
		if !added {
			break
		}
	}
	return out
}

func rowIndexByName(name string) int {
	mu.Lock()
	defer mu.Unlock()
	for i := range rows {
		if rows[i].Name == name {
			return i
		}
	}
	return -1
}

func setRunStatus(runID uint32, text string) {
	mu.Lock()
	if activeRunID != runID {
		mu.Unlock()
		return
	}
	currentStatus = text
	mu.Unlock()
	procPostMessageW.Call(uintptr(hwndMain), WM_APP_STATUS, uintptr(runID), 0)
}

func updateRow(index int, fn func(*resultRow)) {
	mu.Lock()
	if index >= 0 && index < len(rows) {
		fn(&rows[index])
	}
	mu.Unlock()
	procPostMessageW.Call(uintptr(hwndMain), WM_APP_REFRESH, 0, 0)
}

func directSwitchSelected() {
	mu.Lock()
	if running || manualSwitching {
		mu.Unlock()
		return
	}
	ep := activeEndpoint
	mu.Unlock()
	if ep.Base == "" && ep.Pipe == "" {
		msgBox(appTitle, "尚未连接 Clash/Mihomo。", MB_OK|MB_ICONWARNING)
		return
	}
	row, _, ok := selectedResultRow()
	if !ok || row.Name == "" {
		msgBox(appTitle, "请先在结果表中选中一个节点。", MB_OK|MB_ICONWARNING)
		return
	}
	group := selectedGroup()
	if group == "" {
		msgBox(appTitle, "请选择一个策略组。", MB_OK|MB_ICONWARNING)
		return
	}

	mu.Lock()
	manualSwitching = true
	switchTarget = row.Name
	switchGroupName = group
	switchResult = ""
	switchSucceeded = false
	mu.Unlock()
	procEnableWindow.Call(uintptr(hStart), 0)
	procEnableWindow.Call(uintptr(hRetest), 0)
	procEnableWindow.Call(uintptr(hGroup), 0)
	setText(hStatus, fmt.Sprintf("正在切换 %s → %s ...", group, row.Name))

	go func(node string) {
		ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
		err := switchGroup(ctx, ep, group, node)
		cancel()
		mu.Lock()
		if err != nil {
			switchResult = "切换失败：" + err.Error()
			switchSucceeded = false
		} else {
			switchResult = fmt.Sprintf("已切换：%s → %s", group, node)
			switchSucceeded = true
		}
		mu.Unlock()
		procPostMessageW.Call(uintptr(hwndMain), WM_APP_SWITCH, 0, 0)
	}(row.Name)
}

func startSelectedRetest() {
	mu.Lock()
	isRun := running
	ep := activeEndpoint
	fallbackPort := detectedProxyPort
	mu.Unlock()
	if isRun {
		return
	}
	row, rowIdx, ok := selectedResultRow()
	if !ok || row.Name == "" {
		msgBox(appTitle, "请先在结果表中选中一个节点。", MB_OK|MB_ICONWARNING)
		return
	}
	if ep.Base == "" && ep.Pipe == "" {
		msgBox(appTitle, "尚未连接 Clash/Mihomo。", MB_OK|MB_ICONWARNING)
		return
	}
	group := selectedGroup()
	if group == "" {
		msgBox(appTitle, "请选择一个策略组。", MB_OK|MB_ICONWARNING)
		return
	}
	profile := selectedProfile()

	mu.Lock()
	gp, ok := loadedProxies[group]
	if !ok {
		mu.Unlock()
		msgBox(appTitle, "策略组信息已失效，请重新检测。", MB_OK|MB_ICONERROR)
		return
	}
	original := gp.Now
	mu.Unlock()
	if original == "" {
		msgBox(appTitle, "无法确定当前节点，请重新检测后再试。", MB_OK|MB_ICONWARNING)
		return
	}

	if err := saveRecoveryState(group, original); err != nil {
		msgBox(appTitle, "无法创建异常退出恢复记录："+err.Error(), MB_OK|MB_ICONWARNING)
	}

	ctx, cancel := context.WithCancel(context.Background())
	mu.Lock()
	cancelRun = cancel
	running = true
	singleRetest = true
	lastError = ""
	runSeq++
	runID := runSeq
	activeRunID = runID
	currentStatus = ""
	mu.Unlock()
	setRunning(true)
	updateRow(rowIdx, func(r *resultRow) {
		r.Latency = 0
		r.Speed = 0
		r.P90 = 0
		r.Peak = 0
		r.Route = ""
		r.Status = "重测准备"
	})
	setRunStatus(runID, "正在重测："+row.Name)

	go func(node string, idx int) {
		outcome := uintptr(0)
		defer func() {
			restoreCtx, cancelRestore := context.WithTimeout(context.Background(), 8*time.Second)
			restoreErr := switchGroup(restoreCtx, ep, group, original)
			cancelRestore()
			if restoreErr == nil {
				setLocalGroupNow(group, original)
				clearRecoveryState()
			} else {
				mu.Lock()
				if lastError == "" {
					lastError = "恢复原节点失败：" + restoreErr.Error()
				}
				mu.Unlock()
				outcome = 2
			}
			if ctx.Err() != nil && outcome == 0 {
				outcome = 1
			}
			mu.Lock()
			running = false
			cancelRun = nil
			mu.Unlock()
			procPostMessageW.Call(uintptr(hwndMain), WM_APP_DONE, uintptr(runID), outcome)
		}()

		proxyPort, err := getProxyPort(ctx, ep, fallbackPort)
		if err != nil {
			mu.Lock()
			lastError = err.Error()
			mu.Unlock()
			outcome = 2
			return
		}
		updateRow(idx, func(r *resultRow) { r.Status = "延迟扫描" })
		latency := measureLatency(ctx, ep, node)
		updateRow(idx, func(r *resultRow) {
			r.Latency = latency
			if latency > 0 {
				r.Status = "切换中"
			} else {
				r.Status = "延迟失败"
			}
		})
		if ctx.Err() != nil {
			outcome = 1
			return
		}
		if err := switchGroup(ctx, ep, group, node); err != nil {
			updateRow(idx, func(r *resultRow) { r.Status = "切换失败" })
			return
		}
		setLocalGroupNow(group, node)
		select {
		case <-ctx.Done():
			outcome = 1
			return
		case <-time.After(profile.Settle):
		}
		updateRow(idx, func(r *resultRow) { r.Status = "测速中" })
		sr, err := measureSpeed(ctx, ep, node, proxyPort, profile)
		if err != nil {
			if ctx.Err() != nil {
				outcome = 1
				return
			}
			updateRow(idx, func(r *resultRow) {
				r.Route = sr.Route
				if sr.RouteFound && !sr.RouteOK {
					r.Status = "分流异常"
				} else {
					r.Status = "测速失败"
				}
			})
			return
		}
		updateRow(idx, func(r *resultRow) {
			r.Speed = sr.Average
			r.P90 = sr.P90
			r.Peak = sr.Peak
			r.Route = sr.Route
			r.Status = "完成"
		})
		setRunStatus(runID, fmt.Sprintf("重测完成：%s · 平均 %.2f MB/s", node, sr.Average))
	}(row.Name, rowIdx)
}

func startTest() {
	mu.Lock()
	isRun := running
	ep := activeEndpoint
	fallbackPort := detectedProxyPort
	mu.Unlock()
	if isRun {
		return
	}
	if ep.Base == "" && ep.Pipe == "" {
		msgBox(appTitle, "尚未连接 Clash/Mihomo。请先等待自动检测，或在“高级”中手动连接。", MB_OK|MB_ICONWARNING)
		return
	}
	group := selectedGroup()
	if group == "" {
		msgBox(appTitle, "请选择一个策略组。", MB_OK|MB_ICONWARNING)
		return
	}
	profile := selectedProfile()
	scope := selectedScope()
	perRegion := selectedPerRegion()

	mu.Lock()
	pmap := loadedProxies
	gp, ok := pmap[group]
	if !ok {
		mu.Unlock()
		msgBox(appTitle, "策略组信息已失效，请重新检测。", MB_OK|MB_ICONERROR)
		return
	}
	original := gp.Now
	var allNodes []string
	for _, name := range gp.All {
		if p, exists := pmap[name]; exists && isLeafProxy(p) {
			allNodes = append(allNodes, name)
		}
	}
	mu.Unlock()
	if len(allNodes) == 0 {
		msgBox(appTitle, "这个策略组里没有可直接测速的节点。", MB_OK|MB_ICONWARNING)
		return
	}
	nodes, scopeErr := targetNodesForScope(allNodes)
	if scopeErr != "" {
		msgBox(appTitle, scopeErr, MB_OK|MB_ICONWARNING)
		return
	}

	scopeDesc := scopeDescriptionForRun(nodes, perRegion)
	warning := fmt.Sprintf("即将开始测速。\r\n\r\n范围：%s\r\n策略组：%s\r\n测速前节点：%s\r\n\r\n第一阶段的延迟扫描不会切换策略组；第二阶段带宽测速会临时切换该策略组中的节点，因此正在使用同一策略组的程序可能短暂断线。\r\n\r\n本工具不会修改 Windows 系统代理、TUN、DNS、路由表或网卡设置；测速结束/停止/正常关闭时会恢复原节点。\r\n\r\n继续吗？", scopeDesc, group, original)
	if msgBoxResult(appTitle, warning, MB_YESNO|MB_ICONWARNING) != IDYES {
		return
	}

	if err := saveRecoveryState(group, original); err != nil {
		msgBox(appTitle, "无法创建异常退出恢复记录："+err.Error()+"\r\n\r\n本次测速仍可继续，但如果进程被强制结束，程序将无法在下次启动时提示恢复原节点。", MB_OK|MB_ICONWARNING)
	}

	mu.Lock()
	rows = make([]resultRow, 0, len(nodes))
	for _, n := range nodes {
		r := detectRegion(n)
		rows = append(rows, resultRow{Name: n, Region: r, Line: detectLine(n, r), Type: pmap[n].Type, Status: "等待延迟"})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancelRun = cancel
	running = true
	singleRetest = false
	lastError = ""
	runSeq++
	runID := runSeq
	activeRunID = runID
	currentStatus = ""
	mu.Unlock()
	refreshList()
	setRunning(true)
	setRunStatus(runID, fmt.Sprintf("第一阶段：准备延迟扫描 0/%d；这一阶段不会切换当前策略组。", len(nodes)))

	go func() {
		outcome := uintptr(0) // 0=完成, 1=停止, 2=错误
		defer func() {
			if ctx.Err() != nil && outcome == 0 {
				outcome = 1
			}
			restoreCtx, cancelRestore := context.WithTimeout(context.Background(), 8*time.Second)
			restoreErr := error(nil)
			if original != "" {
				restoreErr = switchGroup(restoreCtx, ep, group, original)
			}
			cancelRestore()
			if restoreErr == nil {
				clearRecoveryState()
			} else {
				mu.Lock()
				if lastError == "" {
					lastError = "恢复原节点失败：" + restoreErr.Error()
				}
				mu.Unlock()
				outcome = 2
			}

			mu.Lock()
			if ctx.Err() != nil {
				for i := range rows {
					switch rows[i].Status {
					case "等待延迟", "延迟扫描", "候选", "切换中", "测速中":
						rows[i].Status = "已停止"
					}
				}
			}
			running = false
			cancelRun = nil
			mu.Unlock()
			procPostMessageW.Call(uintptr(hwndMain), WM_APP_DONE, uintptr(runID), outcome)
		}()

		proxyPort, err := getProxyPort(ctx, ep, fallbackPort)
		if err != nil {
			mu.Lock()
			lastError = err.Error()
			mu.Unlock()
			outcome = 2
			return
		}

		// Stage 1: latency-only scan. This calls Mihomo's per-node delay API and does
		// not switch the user's selected strategy group.
		latencies := make(map[string]int, len(nodes))
		stage1Start := time.Now()
		for i, node := range nodes {
			select {
			case <-ctx.Done():
				outcome = 1
				return
			default:
			}
			idx := rowIndexByName(node)
			if idx >= 0 {
				updateRow(idx, func(r *resultRow) { r.Status = "延迟扫描" })
			}
			latency := measureLatency(ctx, ep, node)
			latencies[node] = latency
			if idx >= 0 {
				updateRow(idx, func(r *resultRow) {
					r.Latency = latency
					if latency > 0 {
						r.Status = "延迟完成"
					} else {
						r.Status = "延迟失败"
					}
				})
			}
			done := i + 1
			elapsed := time.Since(stage1Start)
			remaining := ""
			if done > 0 && done < len(nodes) {
				eta := time.Duration(float64(elapsed) / float64(done) * float64(len(nodes)-done))
				remaining = fmt.Sprintf(" · 预计剩余 %s", shortDuration(eta))
			}
			setRunStatus(runID, fmt.Sprintf("第一阶段：延迟扫描 %d/%d · %s%s", done, len(nodes), node, remaining))
		}

		var candidates []string
		if scope == scopeSmart || (scope == scopeRegions && perRegion > 0) {
			candidates = selectRegionalCandidates(nodes, latencies, perRegion)
		} else {
			candidates = append([]string(nil), nodes...)
		}
		candidateSet := map[string]bool{}
		for _, n := range candidates {
			candidateSet[n] = true
		}
		for _, n := range nodes {
			idx := rowIndexByName(n)
			if idx < 0 {
				continue
			}
			if candidateSet[n] {
				updateRow(idx, func(r *resultRow) { r.Status = "候选" })
			} else {
				updateRow(idx, func(r *resultRow) { r.Status = "仅延迟" })
			}
		}
		if len(candidates) == 0 {
			mu.Lock()
			lastError = "延迟筛选后没有可测速候选节点"
			mu.Unlock()
			outcome = 2
			return
		}

		stage2Start := time.Now()
		setRunStatus(runID, fmt.Sprintf("第二阶段：开始带宽测速 0/%d；将临时切换“%s”，结束后恢复“%s”。", len(candidates), group, original))
		for i, node := range candidates {
			select {
			case <-ctx.Done():
				outcome = 1
				return
			default:
			}
			idx := rowIndexByName(node)
			if idx < 0 {
				continue
			}
			updateRow(idx, func(r *resultRow) { r.Status = "切换中" })
			if err := switchGroup(ctx, ep, group, node); err != nil {
				updateRow(idx, func(r *resultRow) { r.Status = "切换失败" })
				continue
			}
			select {
			case <-ctx.Done():
				outcome = 1
				return
			case <-time.After(profile.Settle):
			}
			updateRow(idx, func(r *resultRow) { r.Status = "测速中" })
			sr, err := measureSpeed(ctx, ep, node, proxyPort, profile)
			if err != nil {
				if ctx.Err() != nil {
					outcome = 1
					return
				}
				updateRow(idx, func(r *resultRow) {
					r.Route = sr.Route
					if sr.RouteFound && !sr.RouteOK {
						r.Status = "分流异常"
					} else {
						r.Status = "测速失败"
					}
				})
			} else {
				updateRow(idx, func(r *resultRow) {
					r.Speed = sr.Average
					r.P90 = sr.P90
					r.Peak = sr.Peak
					r.Route = sr.Route
					r.Status = "完成"
				})
			}

			done := i + 1
			elapsed := time.Since(stage2Start)
			remaining := ""
			if done > 0 && done < len(candidates) {
				eta := time.Duration(float64(elapsed) / float64(done) * float64(len(candidates)-done))
				remaining = fmt.Sprintf(" · 预计剩余 %s", shortDuration(eta))
			}
			setRunStatus(runID, fmt.Sprintf("第二阶段：已完成 %d/%d · %s%s", done, len(candidates), node, remaining))
		}

		mu.Lock()
		if scope == scopeSmart || scope == scopeRegions {
			// Keep regions visually grouped; within each region put measured fast nodes first.
			regionRank := map[string]int{}
			rank := 0
			for _, n := range nodes {
				r := detectRegion(n)
				if _, ok := regionRank[r]; !ok {
					regionRank[r] = rank
					rank++
				}
			}
			sort.SliceStable(rows, func(i, j int) bool {
				ri, rj := regionRank[rows[i].Region], regionRank[rows[j].Region]
				if ri != rj {
					return ri < rj
				}
				if rows[i].Speed != rows[j].Speed {
					return rows[i].Speed > rows[j].Speed
				}
				return rows[i].Latency > 0 && (rows[j].Latency == 0 || rows[i].Latency < rows[j].Latency)
			})
		} else {
			sort.SliceStable(rows, func(i, j int) bool {
				if rows[i].Speed == rows[j].Speed {
					return rows[i].Latency > 0 && (rows[j].Latency == 0 || rows[i].Latency < rows[j].Latency)
				}
				return rows[i].Speed > rows[j].Speed
			})
		}
		mu.Unlock()
	}()
}

func shortDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	sec := int(d.Round(time.Second).Seconds())
	if sec < 60 {
		return fmt.Sprintf("%d 秒", sec)
	}
	return fmt.Sprintf("%d:%02d", sec/60, sec%60)
}

func stopTest() {
	mu.Lock()
	c := cancelRun
	runID := activeRunID
	mu.Unlock()
	if c != nil {
		c()
		setRunStatus(runID, "正在停止，并恢复原节点 ...")
	}
}

func regionWndProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_COMMAND:
		id := int32(loword(wParam))
		switch id {
		case IDC_REGION_OK:
			newSel := map[string]bool{}
			for cid, h := range regionCheckHandles {
				if send(h, BM_GETCHECK, 0, 0) == BST_CHECKED {
					if name := regionCheckNames[cid]; name != "" {
						newSel[name] = true
					}
				}
			}
			selectedRegions = newSel
			procDestroyWindow.Call(uintptr(hwnd))
			return 0
		case IDC_REGION_CANCEL:
			procDestroyWindow.Call(uintptr(hwnd))
			return 0
		}
	case WM_APP_WAKE:
		procShowWindow.Call(uintptr(hwndMain), SW_RESTORE)
		procShowWindow.Call(uintptr(hwndMain), SW_SHOW)
		procSetForegroundWindow.Call(uintptr(hwndMain))
		return 0
	case WM_CLOSE:
		procDestroyWindow.Call(uintptr(hwnd))
		return 0
	case WM_DESTROY:
		hwndRegionPicker = 0
		regionCheckHandles = map[int]syscall.Handle{}
		regionCheckNames = map[int]string{}
		procEnableWindow.Call(uintptr(hwndMain), 1)
		updateRegionButton()
		updateEstimate()
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return r
}

func showRegionPicker() {
	if hwndRegionPicker != 0 {
		return
	}
	regions := orderedAvailableRegions()
	if len(regions) == 0 {
		msgBox(appTitle, "当前策略组没有可识别的地区节点。", MB_OK|MB_ICONWARNING)
		return
	}
	counts := availableRegionCounts()
	hInstRaw, _, _ := procGetModuleHandleW.Call(0)
	hInst := syscall.Handle(hInstRaw)
	className := u16("ClashBandwidthTestRegionPicker")
	cur, _, _ := procLoadCursorW.Call(0, IDC_ARROW)
	appIcon, _, _ := procLoadIconW.Call(uintptr(hInst), 2)
	wc := WNDCLASSEX{
		CbSize: uint32(unsafe.Sizeof(WNDCLASSEX{})), LpfnWndProc: syscall.NewCallback(regionWndProc), HInstance: hInst,
		HIcon: syscall.Handle(appIcon), HIconSm: syscall.Handle(appIcon), HCursor: syscall.Handle(cur), HbrBackground: syscall.Handle(COLOR_WINDOW + 1), LpszClassName: className,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))) // already-registered is harmless here
	height := int32(118 + len(regions)*27)
	if height < 260 {
		height = 260
	}
	if height > 610 {
		height = 610
	}
	r, _, _ := procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(u16("选择测速地区"))),
		WS_OVERLAPPED|WS_CAPTION|WS_SYSMENU|WS_VISIBLE,
		CW_USEDEFAULT, CW_USEDEFAULT, 330, uintptr(height), uintptr(hwndMain), 0, uintptr(hInst), 0)
	hwndRegionPicker = syscall.Handle(r)
	if hwndRegionPicker == 0 {
		return
	}
	procEnableWindow.Call(uintptr(hwndMain), 0)
	createControlOn(hwndRegionPicker, 0, "STATIC", "点击勾选一个或多个地区：", WS_CHILD|WS_VISIBLE, 18, 14, 280, 22, 0)
	regionCheckHandles = map[int]syscall.Handle{}
	regionCheckNames = map[int]string{}
	y := int32(42)
	for i, region := range regions {
		id := IDC_REGION_BASE + i
		label := fmt.Sprintf("%s  (%d 个节点)", region, counts[region])
		h := createControlOn(hwndRegionPicker, 0, "BUTTON", label, WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_AUTOCHECKBOX, 22, y, 270, 22, int32(id))
		if selectedRegions[region] {
			send(h, BM_SETCHECK, BST_CHECKED, 0)
		}
		regionCheckHandles[id] = h
		regionCheckNames[id] = region
		y += 27
	}
	buttonY := height - 78
	createControlOn(hwndRegionPicker, 0, "BUTTON", "确定", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 82, buttonY, 75, 28, IDC_REGION_OK)
	createControlOn(hwndRegionPicker, 0, "BUTTON", "取消", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 172, buttonY, 75, 28, IDC_REGION_CANCEL)
	procShowWindow.Call(uintptr(hwndRegionPicker), SW_SHOW)
	procUpdateWindow.Call(uintptr(hwndRegionPicker))
}

func settingsWndProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_COMMAND:
		id := int32(loword(wParam))
		switch id {
		case IDC_SET_SAVE:
			candidate := appSettings{
				CloseToTray:    send(hSetCloseTray, BM_GETCHECK, 0, 0) == BST_CHECKED,
				StartMinimized: send(hSetStartTray, BM_GETCHECK, 0, 0) == BST_CHECKED,
				HotkeyEnabled:  send(hSetHotkey, BM_GETCHECK, 0, 0) == BST_CHECKED,
				Hotkey:         strings.TrimSpace(getText(hSetHotkeyEdit)),
			}
			if candidate.HotkeyEnabled {
				_, _, norm, err := parseHotkey(candidate.Hotkey)
				if err != nil {
					msgBox(appTitle, "快捷键格式无效："+err.Error()+"\r\n\r\n示例：Ctrl+Alt+B", MB_OK|MB_ICONWARNING)
					return 0
				}
				candidate.Hotkey = norm
			}
			old := settings
			if err := registerHotkey(candidate); err != nil {
				_ = registerHotkey(old)
				msgBox(appTitle, "无法启用这个全局快捷键：\r\n"+err.Error()+"\r\n\r\n请换一个组合键。", MB_OK|MB_ICONWARNING)
				return 0
			}
			if err := saveSettings(candidate); err != nil {
				_ = registerHotkey(old)
				msgBox(appTitle, "保存设置失败："+err.Error(), MB_OK|MB_ICONERROR)
				return 0
			}
			settings = candidate
			setText(hStatus, "设置已保存。全局快捷键："+func() string {
				if settings.HotkeyEnabled {
					return settings.Hotkey
				}
				return "已关闭"
			}())
			procDestroyWindow.Call(uintptr(hwnd))
			return 0
		case IDC_SET_CANCEL:
			procDestroyWindow.Call(uintptr(hwnd))
			return 0
		case IDC_SET_HOTKEY:
			enabled := send(hSetHotkey, BM_GETCHECK, 0, 0) == BST_CHECKED
			if enabled {
				procEnableWindow.Call(uintptr(hSetHotkeyEdit), 1)
			} else {
				procEnableWindow.Call(uintptr(hSetHotkeyEdit), 0)
			}
		}
	case WM_APP_WAKE:
		procShowWindow.Call(uintptr(hwndMain), SW_RESTORE)
		procShowWindow.Call(uintptr(hwndMain), SW_SHOW)
		procSetForegroundWindow.Call(uintptr(hwndMain))
		return 0
	case WM_CLOSE:
		procDestroyWindow.Call(uintptr(hwnd))
		return 0
	case WM_DESTROY:
		hwndSettings = 0
		hSetCloseTray = 0
		hSetStartTray = 0
		hSetHotkey = 0
		hSetHotkeyEdit = 0
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return r
}

func showSettingsWindow() {
	if hwndSettings != 0 {
		procShowWindow.Call(uintptr(hwndSettings), SW_RESTORE)
		procSetForegroundWindow.Call(uintptr(hwndSettings))
		return
	}
	hInstRaw, _, _ := procGetModuleHandleW.Call(0)
	hInst := syscall.Handle(hInstRaw)
	className := u16("ClashBandwidthTestSettings")
	cur, _, _ := procLoadCursorW.Call(0, IDC_ARROW)
	wc := WNDCLASSEX{
		CbSize: uint32(unsafe.Sizeof(WNDCLASSEX{})), LpfnWndProc: syscall.NewCallback(settingsWndProc), HInstance: hInst,
		HIcon: appIconHandle, HIconSm: appIconHandle, HCursor: syscall.Handle(cur), HbrBackground: syscall.Handle(COLOR_WINDOW + 1), LpszClassName: className,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	r, _, _ := procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(u16("ClashBandwidthTest 设置"))),
		WS_OVERLAPPED|WS_CAPTION|WS_SYSMENU,
		CW_USEDEFAULT, CW_USEDEFAULT, 490, 330, uintptr(hwndMain), 0, uintptr(hInst), 0)
	hwndSettings = syscall.Handle(r)
	if hwndSettings == 0 {
		return
	}
	createControlOn(hwndSettings, 0, "STATIC", "托盘行为", WS_CHILD|WS_VISIBLE, 20, 18, 100, 22, 0)
	hSetCloseTray = createControlOn(hwndSettings, 0, "BUTTON", "点击关闭按钮 × 时缩小到系统托盘", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_AUTOCHECKBOX, 28, 48, 350, 24, IDC_SET_CLOSETRAY)
	hSetStartTray = createControlOn(hwndSettings, 0, "BUTTON", "启动时直接隐藏到系统托盘", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_AUTOCHECKBOX, 28, 78, 350, 24, IDC_SET_STARTTRAY)
	createControlOn(hwndSettings, 0, "STATIC", "全局快捷键", WS_CHILD|WS_VISIBLE, 20, 122, 100, 22, 0)
	hSetHotkey = createControlOn(hwndSettings, 0, "BUTTON", "启用显示 / 隐藏主窗口快捷键", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_AUTOCHECKBOX, 28, 151, 310, 24, IDC_SET_HOTKEY)
	createControlOn(hwndSettings, 0, "STATIC", "快捷键", WS_CHILD|WS_VISIBLE, 48, 188, 55, 22, 0)
	hSetHotkeyEdit = createControlOn(hwndSettings, WS_EX_CLIENTEDGE, "EDIT", settings.Hotkey, WS_CHILD|WS_VISIBLE|WS_TABSTOP, 107, 184, 180, 25, IDC_SET_HOTKEYEDIT)
	createControlOn(hwndSettings, 0, "STATIC", "格式示例：Ctrl+Alt+B；支持 A-Z、0-9、F1-F12。", WS_CHILD|WS_VISIBLE, 48, 218, 390, 22, 0)
	createControlOn(hwndSettings, 0, "BUTTON", "保存", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 285, 258, 78, 29, IDC_SET_SAVE)
	createControlOn(hwndSettings, 0, "BUTTON", "取消", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 374, 258, 78, 29, IDC_SET_CANCEL)
	if settings.CloseToTray {
		send(hSetCloseTray, BM_SETCHECK, BST_CHECKED, 0)
	}
	if settings.StartMinimized {
		send(hSetStartTray, BM_SETCHECK, BST_CHECKED, 0)
	}
	if settings.HotkeyEnabled {
		send(hSetHotkey, BM_SETCHECK, BST_CHECKED, 0)
		procEnableWindow.Call(uintptr(hSetHotkeyEdit), 1)
	} else {
		procEnableWindow.Call(uintptr(hSetHotkeyEdit), 0)
	}
	procShowWindow.Call(uintptr(hwndSettings), SW_SHOW)
	procUpdateWindow.Call(uintptr(hwndSettings))
}

func showTrayMenu() {
	menu, _, _ := procCreatePopupMenu.Call()
	if menu == 0 {
		return
	}
	defer procDestroyMenu.Call(menu)
	procAppendMenuW.Call(menu, MF_STRING, ID_TRAY_SHOW, uintptr(unsafe.Pointer(u16("显示主窗口"))))
	procAppendMenuW.Call(menu, MF_STRING, ID_TRAY_SETTINGS, uintptr(unsafe.Pointer(u16("设置..."))))
	procAppendMenuW.Call(menu, MF_SEPARATOR, 0, 0)
	procAppendMenuW.Call(menu, MF_STRING, ID_TRAY_EXIT, uintptr(unsafe.Pointer(u16("退出"))))
	var pt POINT
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	procSetForegroundWindow.Call(uintptr(hwndMain))
	cmd, _, _ := procTrackPopupMenu.Call(menu, TPM_RIGHTBUTTON|TPM_RETURNCMD, uintptr(pt.X), uintptr(pt.Y), 0, uintptr(hwndMain), 0)
	procPostMessageW.Call(uintptr(hwndMain), 0, 0, 0)
	switch cmd {
	case ID_TRAY_SHOW:
		restoreMainWindow()
	case ID_TRAY_SETTINGS:
		restoreMainWindow()
		showSettingsWindow()
	case ID_TRAY_EXIT:
		requestExit()
	}
}

func listWndProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	if msg == WM_LBUTTONDBLCLK {
		procPostMessageW.Call(uintptr(hwndMain), WM_APP_LISTDBL, 0, 0)
	}
	if oldListWndProc != 0 {
		r, _, _ := procCallWindowProcW.Call(oldListWndProc, uintptr(hwnd), uintptr(msg), wParam, lParam)
		return r
	}
	r, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return r
}

var resizingWindow bool

func rectFromDPIParam(p uintptr) [4]int32 {
	var rect [4]int32
	if p != 0 {
		procRtlMoveMemory.Call(
			uintptr(unsafe.Pointer(&rect[0])),
			p,
			uintptr(len(rect)*4),
		)
	}
	return rect
}

func wndProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_ENTERSIZEMOVE:
		resizingWindow = true
		send(hList, WM_SETREDRAW, 0, 0)
		return 0
	case WM_EXITSIZEMOVE:
		resizingWindow = false
		send(hList, WM_SETREDRAW, 1, 0)
		resizeListColumns()
		scheduleListRebuild()
		return 0
	case WM_DPICHANGED:
		// Apply Windows recommended rectangle before rebuilding controls.
		// The old implementation refreshed ListView while DPI transition was
		// still in progress, which could leave rows blank after monitor changes.
		if lParam != 0 {
			r := rectFromDPIParam(lParam)
			procSetWindowPos.Call(uintptr(hwnd), 0,
				uintptr(r[0]), uintptr(r[1]),
				uintptr(r[2]-r[0]), uintptr(r[3]-r[1]),
				SWP_NOZORDER|SWP_NOACTIVATE)
		}
		resizeListColumns()
		scheduleListRebuild()
		return 0
	case WM_SIZE:
		if !resizingWindow {
			resizeListColumns()
			scheduleListRebuild()
		}
		return 0
	case WM_COMMAND:
		id := int32(loword(wParam))
		notify := hiword(wParam)
		switch id {
		case IDC_CONNECT:
			connectAsync()
		case IDC_REDETECT:
			autoDetectAsync()
		case IDC_ADVANCED:
			setAdvancedVisible(!advancedVisible)
		case IDC_SETTINGS:
			showSettingsWindow()
		case IDC_START:
			startTest()
		case IDC_RETEST:
			startSelectedRetest()
		case IDC_STOP:
			stopTest()
		case IDC_REGIONS:
			showRegionPicker()
		case IDC_FILTER:
			if notify == EN_CHANGE {
				updateEstimate()
			}
		case IDC_GROUP:
			if notify == CBN_SELCHANGE {
				updateRegionButton()
				updateEstimate()
			}
		case IDC_SIZE, IDC_PERREGION:
			if notify == CBN_SELCHANGE {
				updateEstimate()
			}
		case IDC_SCOPE:
			if notify == CBN_SELCHANGE {
				updateScopeControls()
				updateEstimate()
			}
		}
		return 0
	case WM_HOTKEY:
		if wParam == HOTKEY_ID {
			toggleMainWindow()
		}
		return 0
	case WM_APP_TRAY:
		trayMsg := uint32(lParam & 0xffff)
		switch trayMsg {
		case WM_LBUTTONDBLCLK:
			restoreMainWindow()
		case WM_RBUTTONUP:
			showTrayMenu()
		}
		return 0
	case WM_APP_LISTDBL:
		directSwitchSelected()
		return 0
	case WM_APP_GROUPS:
		procEnableWindow.Call(uintptr(hConnect), 1)
		mu.Lock()
		errText := lastError
		n := len(groupNames)
		ep := activeEndpoint
		mu.Unlock()
		if errText != "" {
			setText(hStatus, "手动连接失败")
			hint := errText
			if strings.Contains(errText, "401") {
				hint += "\r\n\r\n请检查 Core Secret。"
			}
			msgBox(appTitle, hint, MB_OK|MB_ICONERROR)
		} else {
			populateGroups()
			updateScopeControls()
			updateRegionButton()
			updateEstimate()
			procEnableWindow.Call(uintptr(hGroup), 1)
			procEnableWindow.Call(uintptr(hStart), 1)
			setText(hDetected, "● 已连接 · "+ep.Mode)
			setText(hStatus, fmt.Sprintf("已连接，发现 %d 个 Selector 策略组。", n))
			maybeOfferRecovery()
		}
		return 0
	case WM_APP_DETECT:
		procEnableWindow.Call(uintptr(hRedetect), 1)
		mu.Lock()
		errText := lastError
		n := len(groupNames)
		ep := activeEndpoint
		pport := detectedProxyPort
		cfgPath := detectedConfigPath
		mu.Unlock()
		if errText != "" {
			setText(hDetected, "○ 未自动连接 Clash Verge Rev")
			setText(hStatus, "自动检测失败。已展开高级设置，可手动填写 Controller / Secret，或点击“重新检测”。")
			setAdvancedVisible(true)
			return 0
		}
		populateGroups()
		updateScopeControls()
		updateRegionButton()
		updateEstimate()
		procEnableWindow.Call(uintptr(hGroup), 1)
		procEnableWindow.Call(uintptr(hStart), 1)
		if ep.Base != "" {
			setText(hController, ep.Base)
		}
		if ep.Secret != "" {
			setText(hSecret, ep.Secret)
		}
		info := "● Clash Verge Rev 已连接 · Mihomo · " + ep.Mode
		if pport > 0 {
			info += fmt.Sprintf(" · Mixed Port %d", pport)
		}
		setText(hDetected, info)
		status := fmt.Sprintf("自动检测完成：发现 %d 个 Selector 策略组。", n)
		if cfgPath != "" {
			status += " 已读取本机运行配置。"
		}
		setText(hStatus, status)
		setAdvancedVisible(false)
		maybeOfferRecovery()
		return 0
	case WM_APP_REFRESH:
		refreshList()
		return 0
	case WM_APP_STATUS:
		mu.Lock()
		active := activeRunID
		status := currentStatus
		mu.Unlock()
		if uint32(wParam) == active && active != 0 {
			setText(hStatus, status)
		}
		return 0
	case WM_APP_SWITCH:
		mu.Lock()
		text := switchResult
		target := switchTarget
		group := switchGroupName
		ok := switchSucceeded
		manualSwitching = false
		switchResult = ""
		switchGroupName = ""
		mu.Unlock()
		if ok {
			setLocalGroupNow(group, target)
			setText(hStatus, text+"。双击其他节点可继续切换。")
		} else {
			setText(hStatus, text)
			msgBox(appTitle, text, MB_OK|MB_ICONERROR)
		}
		mu.Lock()
		hasRows := len(rows) > 0
		mu.Unlock()
		procEnableWindow.Call(uintptr(hStart), 1)
		procEnableWindow.Call(uintptr(hGroup), 1)
		if hasRows {
			procEnableWindow.Call(uintptr(hRetest), 1)
		}
		return 0
	case WM_APP_RECOVER:
		mu.Lock()
		text := recoveryResult
		recoveryResult = ""
		mu.Unlock()
		if text != "" {
			setText(hStatus, text)
		}
		return 0
	case WM_APP_DONE:
		mu.Lock()
		if uint32(wParam) != activeRunID || activeRunID == 0 {
			mu.Unlock()
			return 0
		}
		errText := lastError
		wasSingle := singleRetest
		singleRetest = false
		shouldClose := closePending
		closePending = false
		activeRunID = 0
		mu.Unlock()
		refreshList()
		setRunning(false)
		if shouldClose {
			procDestroyWindow.Call(uintptr(hwndMain))
			return 0
		}
		switch lParam {
		case 1:
			setText(hStatus, "测速已停止，并已尝试恢复测速前节点。")
		case 2:
			setText(hStatus, "测速结束，但出现错误；原节点已尝试恢复。")
			if errText != "" {
				msgBox(appTitle, errText, MB_OK|MB_ICONERROR)
			}
		default:
			if wasSingle {
				setText(hStatus, "选中节点重测完成，并已恢复重测前节点。")
			} else {
				setText(hStatus, "测速完成：结果已整理，并已恢复测速前节点。")
			}
		}
		return 0
	case WM_APP_WAKE:
		procShowWindow.Call(uintptr(hwndMain), SW_RESTORE)
		procShowWindow.Call(uintptr(hwndMain), SW_SHOW)
		procSetForegroundWindow.Call(uintptr(hwndMain))
		return 0
	case WM_CLOSE:
		if settings.CloseToTray {
			procShowWindow.Call(uintptr(hwndMain), SW_HIDE)
			return 0
		}
		requestExit()
		return 0
	case WM_DESTROY:
		unregisterHotkey()
		removeTrayIcon()
		if hwndSettings != 0 {
			procDestroyWindow.Call(uintptr(hwndSettings))
		}
		procPostQuitMessage.Call(0)
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return r
}


var instanceMutex syscall.Handle

// ensureSingleInstance prevents multiple GUI instances from running.
// A second launch simply exits; the existing window remains available.
func ensureSingleInstance() bool {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	createMutex := kernel32.NewProc("CreateMutexW")

	name, _ := syscall.UTF16PtrFromString("Global\\ClashBandwidthTest_Instance_v1")
	h, _, _ := createMutex.Call(0, 0, uintptr(unsafe.Pointer(name)))
	if h == 0 {
		return false
	}
	instanceMutex = syscall.Handle(h)

	err := syscall.GetLastError()
	if err == syscall.ERROR_ALREADY_EXISTS {
		className, _ := syscall.UTF16PtrFromString("ClashBandwidthTestWnd")
		if hwnd, _, _ := procFindWindowW.Call(uintptr(unsafe.Pointer(className)), 0); hwnd != 0 {
			procPostMessageW.Call(hwnd, WM_APP_WAKE, 0, 0)
		}
		return false
	}
	return true
}

func main() {
	// A Win32 window and its message queue are bound to the OS thread that
	// created them. Keep the full GUI lifetime on one OS thread.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if !ensureSingleInstance() {
		return
	}

	settings = loadSettings()

	icc := INITCOMMONCONTROLSEX{DwSize: uint32(unsafe.Sizeof(INITCOMMONCONTROLSEX{})), DwICC: ICC_LISTVIEW_CLASSES}
	procInitCommonControlsEx.Call(uintptr(unsafe.Pointer(&icc)))
	hInstRaw, _, _ := procGetModuleHandleW.Call(0)
	hInst := syscall.Handle(hInstRaw)
	cur, _, _ := procLoadCursorW.Call(0, IDC_ARROW)
	// Resource IDs: manifest=1, icon group=2 (generated by rsrc).
	appIcon, _, _ := procLoadIconW.Call(uintptr(hInst), 2)
	appIconHandle = syscall.Handle(appIcon)
	f, _, _ := procGetStockObject.Call(DEFAULT_GUI_FONT)
	guiFont = syscall.Handle(f)
	className := u16("ClashBandwidthTestWnd")
	wc := WNDCLASSEX{
		CbSize: uint32(unsafe.Sizeof(WNDCLASSEX{})), LpfnWndProc: syscall.NewCallback(wndProc), HInstance: hInst,
		HIcon: syscall.Handle(appIcon), HIconSm: syscall.Handle(appIcon), HCursor: syscall.Handle(cur), HbrBackground: syscall.Handle(COLOR_WINDOW + 1), LpszClassName: className,
	}
	if r, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); r == 0 {
		return
	}
	r, _, _ := procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(u16(appTitle+" 1.8"))),
		WS_OVERLAPPED|WS_CAPTION|WS_SYSMENU|WS_MINIMIZEBOX,
		CW_USEDEFAULT, CW_USEDEFAULT, 1060, 740, 0, 0, uintptr(hInst), 0)
	hwndMain = syscall.Handle(r)
	if hwndMain == 0 {
		return
	}

	// Clean default view: auto-detection status + flexible scope selection.
	hDetected = createControl(0, "STATIC", "● 正在自动检测 Clash Verge Rev ...", WS_CHILD|WS_VISIBLE, 18, 18, 800, 24, 0)
	hSettings = createControl(0, "BUTTON", "设置", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 845, 14, 82, 28, IDC_SETTINGS)
	hAdvanced = createControl(0, "BUTTON", "高级", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 940, 14, 86, 28, IDC_ADVANCED)

	createControl(0, "STATIC", "策略组", WS_CHILD|WS_VISIBLE, 18, 58, 55, 22, 0)
	hGroup = createControl(WS_EX_CLIENTEDGE, "COMBOBOX", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|CBS_DROPDOWNLIST|WS_VSCROLL, 75, 54, 250, 240, IDC_GROUP)
	createControl(0, "STATIC", "范围", WS_CHILD|WS_VISIBLE, 340, 58, 42, 22, 0)
	hScope = createControl(WS_EX_CLIENTEDGE, "COMBOBOX", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|CBS_DROPDOWNLIST, 383, 54, 170, 120, IDC_SCOPE)
	for _, text := range []string{"地区智能抽样", "指定地区", "按名称筛选", "全部节点"} {
		send(hScope, CB_ADDSTRING, 0, uintptr(unsafe.Pointer(u16(text))))
	}
	send(hScope, CB_SETCURSEL, 0, 0)
	hPerRegionLabel = createControl(0, "STATIC", "每地区", WS_CHILD|WS_VISIBLE, 566, 58, 58, 22, 0)
	hPerRegion = createControl(WS_EX_CLIENTEDGE, "COMBOBOX", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|CBS_DROPDOWNLIST, 625, 54, 82, 140, IDC_PERREGION)
	for _, text := range []string{"全部", "1 个", "2 个", "3 个"} {
		send(hPerRegion, CB_ADDSTRING, 0, uintptr(unsafe.Pointer(u16(text))))
	}
	send(hPerRegion, CB_SETCURSEL, 2, 0)
	hStart = createControl(0, "BUTTON", "开始测速", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 730, 53, 102, 28, IDC_START)
	hRetest = createControl(0, "BUTTON", "重测选中", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 838, 53, 102, 28, IDC_RETEST)
	hStop = createControl(0, "BUTTON", "停止", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 946, 53, 70, 28, IDC_STOP)
	procEnableWindow.Call(uintptr(hStop), 0)
	procEnableWindow.Call(uintptr(hRetest), 0)
	procEnableWindow.Call(uintptr(hStart), 0)
	procEnableWindow.Call(uintptr(hGroup), 0)

	createControl(0, "STATIC", "范围条件", WS_CHILD|WS_VISIBLE, 18, 94, 68, 22, 0)
	hScopeDetail = createControl(0, "STATIC", "自动识别全部地区", WS_CHILD|WS_VISIBLE, 88, 94, 262, 22, IDC_SCOPEDETAIL)
	hRegionSelect = createControl(0, "BUTTON", "选择地区…", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 88, 89, 262, 28, IDC_REGIONS)
	hNameFilter = createControl(WS_EX_CLIENTEDGE, "EDIT", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP, 88, 90, 262, 25, IDC_FILTER)
	createControl(0, "STATIC", "测速模式", WS_CHILD|WS_VISIBLE, 370, 94, 68, 22, 0)
	hSize = createControl(WS_EX_CLIENTEDGE, "COMBOBOX", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|CBS_DROPDOWNLIST, 440, 90, 282, 120, IDC_SIZE)
	for _, text := range []string{"快速筛选 · 2秒 / 2路 / ≤13MB", "标准测速 · 5秒 / 4路 / ≤62MB"} {
		send(hSize, CB_ADDSTRING, 0, uintptr(unsafe.Pointer(u16(text))))
	}
	send(hSize, CB_SETCURSEL, 0, 0)
	hEstimate = createControl(0, "STATIC", "等待读取节点数量 ...", WS_CHILD|WS_VISIBLE, 18, 124, 1008, 22, IDC_ESTIMATE)

	hList = createControl(WS_EX_CLIENTEDGE, "SysListView32", "", WS_CHILD|WS_VISIBLE|LVS_REPORT|LVS_SINGLESEL|LVS_SHOWSELALWAYS, 18, 150, 1008, 405, IDC_LIST)
	send(hList, LVM_SETEXTENDEDLISTVIEWSTYLE, 0, LVS_EX_FULLROWSELECT|LVS_EX_GRIDLINES|LVS_EX_DOUBLEBUFFER)
	listWndProcCallback = syscall.NewCallback(listWndProc)
	oldListWndProc, _, _ = procSetWindowLongPtrW.Call(uintptr(hList), ^uintptr(3), listWndProcCallback)
	addColumn(0, "节点", 190)
	addColumn(1, "地区", 62)
	addColumn(2, "线路", 58)
	addColumn(3, "类型", 70)
	addColumn(4, "延迟", 68)
	addColumn(5, "平均速度", 92)
	addColumn(6, "稳态 P90", 88)
	addColumn(7, "峰值速度", 88)
	addColumn(8, "路径验证", 145)
	addColumn(9, "状态", 88)
	resizeListColumns()

	// Advanced/manual fallback. Hidden unless auto-detection fails or the user opens it.
	hConnLabel = createControl(0, "STATIC", "Controller", WS_CHILD|WS_VISIBLE, 18, 580, 75, 22, 0)
	hController = createControl(WS_EX_CLIENTEDGE, "EDIT", "http://127.0.0.1:9090", WS_CHILD|WS_VISIBLE|WS_TABSTOP, 95, 577, 220, 25, IDC_CONTROLLER)
	hSecretLabel = createControl(0, "STATIC", "Secret", WS_CHILD|WS_VISIBLE, 330, 580, 48, 22, 0)
	hSecret = createControl(WS_EX_CLIENTEDGE, "EDIT", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|ES_PASSWORD, 380, 577, 170, 25, IDC_SECRET)
	send(hSecret, EM_SETPASSWORDCHAR, uintptr('*'), 0)
	hConnect = createControl(0, "BUTTON", "手动连接", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 565, 576, 92, 27, IDC_CONNECT)
	hRedetect = createControl(0, "BUTTON", "重新检测", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 668, 576, 92, 27, IDC_REDETECT)

	hStatus = createControl(0, "STATIC", "v1.8：支持系统托盘与全局快捷键。双击节点可直接切换；选中一行可“重测选中”。测速不会修改 Windows 系统代理、TUN、DNS 或路由设置。", WS_CHILD|WS_VISIBLE, 18, 650, 1008, 42, IDC_STATUS)
	setAdvancedVisible(false)
	updateScopeControls()

	addTrayIcon()
	hotkeyErr := registerHotkey(settings)
	if hotkeyErr != nil {
		setText(hStatus, "全局快捷键未启用："+hotkeyErr.Error()+"。可在“设置”中重新选择。")
	}
	if settings.StartMinimized {
		procShowWindow.Call(uintptr(hwndMain), SW_HIDE)
	} else {
		procShowWindow.Call(uintptr(hwndMain), SW_SHOW)
		procUpdateWindow.Call(uintptr(hwndMain))
	}

	if hotkeyErr != nil {
		msgBox(appTitle, "全局快捷键未启用：\r\n"+hotkeyErr.Error()+"\r\n\r\n原设置已保留，请在“设置”中选择其他组合键。", MB_OK|MB_ICONWARNING)
	}

	// Try to attach to Clash Verge Rev automatically. This prefers its private
	// named pipe, then the configured HTTP controller, then common local ports.
	autoDetectAsync()

	var m MSG
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(ret) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}
