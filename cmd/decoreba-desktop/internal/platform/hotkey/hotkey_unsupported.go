//go:build !linux && !windows && !darwin

package hotkey

import "fmt"

func newPlatform(showCh chan<- bool, key string) (*Manager, error) {
	return nil, fmt.Errorf("hotkey not supported on this platform")
}

var keyNames = map[string]struct{}{}
