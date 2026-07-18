package hotkey

import (
	"fmt"
	"sort"
)

type Manager struct {
	closeFn func()
}

func New(showCh chan<- bool) (*Manager, error) {
	return NewKey(showCh, "space")
}

func NewKey(showCh chan<- bool, key string) (*Manager, error) {
	m, err := newPlatform(showCh, key)
	if err != nil {
		return nil, fmt.Errorf("hotkey: %w", err)
	}
	return m, nil
}

func (m *Manager) Close() {
	if m.closeFn != nil {
		m.closeFn()
	}
}

func KnownKeys() string {
	keys := make([]string, 0, len(keyNames))
	for k := range keyNames {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := ""
	for i, k := range keys {
		if i > 0 {
			result += ", "
		}
		result += k
	}
	return result
}
