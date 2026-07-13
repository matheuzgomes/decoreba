//go:build !linux && !windows && !darwin

package tray

import "fmt"

func probePlatform() bool {
	return false
}

func newPlatform(showCh chan<- bool, quitCh chan<- struct{}) (*Tray, error) {
	return nil, fmt.Errorf("tray not supported on this platform")
}
