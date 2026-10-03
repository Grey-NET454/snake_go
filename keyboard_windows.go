//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
	procReadConsoleA   = kernel32.NewProc("ReadConsoleA")
	procGetStdHandle   = kernel32.NewProc("GetStdHandle")

	originalMode uint32
	consoleHandle syscall.Handle
)

const (
	enableRawInputMode = 0x0001 // ENABLE_PROCESSED_INPUT только, без echo и window input
)

// Включаем режим чтения нажатых клавиш без Enter
func enableRawMode() error {
	consoleHandle = syscall.Handle(getStdHandle())
	var mode uint32
	ret, _, err := procGetConsoleMode.Call(uintptr(consoleHandle), uintptr(unsafe.Pointer(&mode)))
	if ret == 0 {
		return err
	}
	originalMode = mode
	// ENABLE_PROCESSED_INPUT = 0x0001 (Ctrl+C обрабатывается)
	// Убираем ENABLE_ECHO_INPUT (0x0004) и ENABLE_LINE_INPUT (0x0002)
	newMode := mode &^ (0x0002 | 0x0004 | 0x0008 | 0x0020)
	newMode |= 0x0001
	ret, _, err = procSetConsoleMode.Call(uintptr(consoleHandle), uintptr(newMode))
	if ret == 0 {
		return err
	}
	return nil
}

// Возвращаем исходный режим консоли
func disableRawMode() {
	procSetConsoleMode.Call(uintptr(consoleHandle), uintptr(originalMode))
}

func getStdHandle() uintptr {
	// STD_INPUT_HANDLE = -10
	ret, _, _ := procGetStdHandle.Call(^uintptr(9)) // -10
	return ret
}

// Чтение нажатых клавиш в канал
func readKeys(ch chan<- byte) {
	buf := make([]byte, 1)
	var read uint32
	for {
		ret, _, _ := procReadConsoleA.Call(
			uintptr(consoleHandle),
			uintptr(unsafe.Pointer(&buf[0])),
			1,
			uintptr(unsafe.Pointer(&read)),
			0,
		)
		if ret == 0 || read == 0 {
			continue
		}
		ch <- buf[0]
	}
}
