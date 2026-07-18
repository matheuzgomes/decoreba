//go:build !linux && !darwin && !windows

package term

import "errors"

var errUnsupportedTerm = errors.New("interactive terminal not supported on this platform")

func IsTerminal() bool             { return false }
func MakeRaw() (func(), error)     { return nil, errUnsupportedTerm }
func GetSize() (width, height int) { return 0, 0 }
func InputAvailable(ms int) bool   { return false }
