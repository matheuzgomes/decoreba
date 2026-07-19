package sync

// Backend persists the command store to a remote location.
type Backend interface {
	Name() string
	Upload(data []byte, remoteID string) (string, error)
	Download(remoteID string) ([]byte, error)
}
