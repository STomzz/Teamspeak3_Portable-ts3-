//go:build windows

package main

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

func showErrorDialog(title, message string) {
	user32 := syscall.NewLazyDLL("user32.dll")
	messageBoxW := user32.NewProc("MessageBoxW")

	titlePtr, _ := syscall.UTF16PtrFromString(title)
	messagePtr, _ := syscall.UTF16PtrFromString(message)

	const mbIconError = 0x00000010
	const mbOK = 0x00000000

	_, _, _ = messageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(messagePtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		mbOK|mbIconError,
	)
}

func launchClient(exePath string, args []string, workDir string, env []string) error {
	shell32 := syscall.NewLazyDLL("shell32.dll")
	shellExecuteExW := shell32.NewProc("ShellExecuteExW")

	verbPtr, _ := syscall.UTF16PtrFromString("open")
	filePtr, _ := syscall.UTF16PtrFromString(exePath)
	paramsPtr, _ := syscall.UTF16PtrFromString(joinWindowsArgs(args))
	dirPtr, _ := syscall.UTF16PtrFromString(workDir)
	restoreEnv, err := applyProcessEnv(env)
	if err != nil {
		return err
	}
	defer restoreEnv()

	info := shellExecuteInfo{
		cbSize:       uint32(unsafe.Sizeof(shellExecuteInfo{})),
		fMask:        seeMaskNoCloseProcess | seeMaskNoAsync | seeMaskFlagDDEWait,
		lpVerb:       verbPtr,
		lpFile:       filePtr,
		lpParameters: paramsPtr,
		lpDirectory:  dirPtr,
		nShow:        swShowDefault,
	}

	r1, _, callErr := shellExecuteExW.Call(uintptr(unsafe.Pointer(&info)))
	if r1 == 0 {
		if callErr != syscall.Errno(0) {
			return fmt.Errorf("ShellExecuteExW failed: %w", callErr)
		}
		return fmt.Errorf("ShellExecuteExW failed")
	}

	return nil
}

const (
	seeMaskNoCloseProcess = 0x00000040
	seeMaskNoAsync        = 0x00000100
	seeMaskFlagDDEWait    = 0x00000100
	seeMaskDoEnvSubst     = 0x00000200
	swShowDefault         = 10
)

type shellExecuteInfo struct {
	cbSize       uint32
	fMask        uint32
	hwnd         uintptr
	lpVerb       *uint16
	lpFile       *uint16
	lpParameters *uint16
	lpDirectory  *uint16
	nShow        int32
	hInstApp     uintptr
	lpIDList     uintptr
	lpClass      *uint16
	hkeyClass    uintptr
	dwHotKey     uint32
	hIconOrMon   uintptr
	hProcess     uintptr
}

func joinWindowsArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}

	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		quoted = append(quoted, quoteWindowsArg(arg))
	}
	return strings.Join(quoted, " ")
}

func quoteWindowsArg(arg string) string {
	if arg == "" {
		return `""`
	}
	if !strings.ContainsAny(arg, " \t\"") {
		return arg
	}

	var b strings.Builder
	b.WriteByte('"')
	backslashes := 0
	for _, r := range arg {
		switch r {
		case '\\':
			backslashes++
		case '"':
			b.WriteString(strings.Repeat(`\`, backslashes*2+1))
			b.WriteRune('"')
			backslashes = 0
		default:
			if backslashes > 0 {
				b.WriteString(strings.Repeat(`\`, backslashes))
				backslashes = 0
			}
			b.WriteRune(r)
		}
	}
	if backslashes > 0 {
		b.WriteString(strings.Repeat(`\`, backslashes*2))
	}
	b.WriteByte('"')
	return b.String()
}

func applyProcessEnv(env []string) (func(), error) {
	type previousValue struct {
		key     string
		value   string
		present bool
	}

	previous := make([]previousValue, 0, len(env))
	for _, entry := range env {
		parts := strings.SplitN(entry, "=", 2)
		key := parts[0]
		if key == "" || strings.HasPrefix(key, "=") {
			continue
		}
		value := ""
		if len(parts) == 2 {
			value = parts[1]
		}

		prev, ok := syscall.Getenv(key)
		previous = append(previous, previousValue{key: key, value: prev, present: ok})
		if err := syscall.Setenv(key, value); err != nil {
			return nil, fmt.Errorf("set env %s: %w", key, err)
		}
	}

	return func() {
		for i := len(previous) - 1; i >= 0; i-- {
			item := previous[i]
			if item.present {
				_ = syscall.Setenv(item.key, item.value)
			} else {
				_ = syscall.Unsetenv(item.key)
			}
		}
	}, nil
}
