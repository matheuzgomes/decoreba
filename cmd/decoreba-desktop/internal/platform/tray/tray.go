package tray

import "fmt"

type Tray struct {
	closeFn func() error
}

func (t *Tray) Close() error {
	if t.closeFn != nil {
		return t.closeFn()
	}
	return nil
}

func Available() bool {
	return probePlatform()
}

func New(showCh chan<- bool, quitCh chan<- struct{}) (*Tray, error) {
	t, err := newPlatform(showCh, quitCh)
	if err != nil {
		return nil, fmt.Errorf("tray: %w", err)
	}
	return t, nil
}
