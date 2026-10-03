//go:build !windows

package main

import (
	"os"
	"os/exec"
)

// Включаем raw-режим через stty (без внешних зависимостей)
func enableRawMode() error {
	cmd := exec.Command("stty", "raw", "-echo")
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Возвращаем исходный режим терминала
func disableRawMode() {
	cmd := exec.Command("stty", "sane")
	cmd.Stdin = os.Stdin
	cmd.Run()
}

// Чтение нажатых клавиш в канал
func readKeys(ch chan<- byte) {
	buf := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			continue
		}
		ch <- buf[0]
	}
}
