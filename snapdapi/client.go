package snapdapi

import (
	"github.com/snapcore/snapd/client"
	"sync"
)

// SnapdClient is a client of the snapd REST API
type SnapdClient interface {
	List(names []string, opts *client.ListOptions) ([]*client.Snap, error)
}

var clientOnce sync.Once
var clientInstance *ClientAdapter

// ClientAdapter adapts our expectations to the snapd client API.
type ClientAdapter struct {
	snapdClient *client.Client
}

// NewClientAdapter creates a new ClientAdapter as a singleton
func NewClientAdapter() *ClientAdapter {
	clientOnce.Do(func() {
		clientInstance = &ClientAdapter{
			snapdClient: client.New(nil),
		}
	})

	return clientInstance
}

// List returns the list of all snaps installed on the system
// with names in the given list; if the list is empty, all snaps.
func (a *ClientAdapter) List(names []string, opts *client.ListOptions) ([]*client.Snap, error) {
	return a.snapdClient.List(names, opts)
}
