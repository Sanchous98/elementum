//go:build !darwin && !freebsd && !dragonfly
// +build !darwin,!freebsd,!dragonfly

package bittorrent

import "os"

func unlockFile() error {
	return nil
}
