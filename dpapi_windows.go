//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

type dataBlob struct {
	Size uint32
	Data *byte
}

var (
	crypt32                = syscall.NewLazyDLL("crypt32.dll")
	procCryptProtectData   = crypt32.NewProc("CryptProtectData")
	procCryptUnprotectData = crypt32.NewProc("CryptUnprotectData")
	procLocalFree          = kernel32.NewProc("LocalFree")
)

const cryptProtectUIForbidden = 0x1

type cachedController struct {
	Base   string `json:"controller"`
	Secret string `json:"secret"`
}

func rememberedEndpointPath() string {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		base, _ = os.UserConfigDir()
	}
	if base == "" {
		return ""
	}
	return filepath.Join(base, "ClashBandwidthTest", "controller.dpapi")
}

func protectCurrentUser(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty DPAPI input")
	}
	input := dataBlob{Size: uint32(len(data)), Data: &data[0]}
	var output dataBlob
	r, _, err := procCryptProtectData.Call(uintptr(unsafe.Pointer(&input)), 0, 0, 0, 0, cryptProtectUIForbidden, uintptr(unsafe.Pointer(&output)))
	if r == 0 {
		return nil, fmt.Errorf("CryptProtectData: %v", err)
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(output.Data)))
	if output.Data == nil {
		return nil, fmt.Errorf("empty DPAPI result")
	}
	return append([]byte(nil), unsafe.Slice(output.Data, int(output.Size))...), nil
}

func unprotectCurrentUser(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty DPAPI input")
	}
	input := dataBlob{Size: uint32(len(data)), Data: &data[0]}
	var output dataBlob
	r, _, err := procCryptUnprotectData.Call(uintptr(unsafe.Pointer(&input)), 0, 0, 0, 0, cryptProtectUIForbidden, uintptr(unsafe.Pointer(&output)))
	if r == 0 {
		return nil, fmt.Errorf("CryptUnprotectData: %v", err)
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(output.Data)))
	if output.Data == nil {
		return nil, fmt.Errorf("empty DPAPI result")
	}
	return append([]byte(nil), unsafe.Slice(output.Data, int(output.Size))...), nil
}

func isLocalController(base string) bool {
	u, err := url.Parse(base)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil {
		return false
	}
	host := u.Hostname()
	return strings.EqualFold(host, "localhost") || (net.ParseIP(host) != nil && net.ParseIP(host).IsLoopback())
}

func saveRememberedEndpoint(ep clashEndpoint) error {
	if !isLocalController(ep.Base) || ep.Pipe != "" {
		return fmt.Errorf("仅允许记住本机 HTTP Controller")
	}
	path := rememberedEndpointPath()
	if path == "" {
		return fmt.Errorf("无法找到本地数据目录")
	}
	raw, err := json.Marshal(cachedController{Base: ep.Base, Secret: ep.Secret})
	if err != nil {
		return err
	}
	encrypted, err := protectCurrentUser(raw)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, encrypted, 0600)
}

func loadRememberedEndpoint() (clashEndpoint, error) {
	path := rememberedEndpointPath()
	if path == "" {
		return clashEndpoint{}, os.ErrNotExist
	}
	encrypted, err := os.ReadFile(path)
	if err != nil {
		return clashEndpoint{}, err
	}
	if len(encrypted) > 32768 {
		return clashEndpoint{}, fmt.Errorf("记忆文件大小异常")
	}
	raw, err := unprotectCurrentUser(encrypted)
	if err != nil {
		return clashEndpoint{}, err
	}
	var cfg cachedController
	if err = json.Unmarshal(raw, &cfg); err != nil {
		return clashEndpoint{}, err
	}
	if !isLocalController(cfg.Base) {
		return clashEndpoint{}, fmt.Errorf("已记住的控制器地址不是本机")
	}
	return clashEndpoint{Base: cfg.Base, Secret: cfg.Secret, Mode: "已记住的 HTTP Controller"}, nil
}

func clearRememberedEndpoint() error {
	p := rememberedEndpointPath()
	if p == "" {
		return nil
	}
	err := os.Remove(p)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
