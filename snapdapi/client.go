package snapdapi

import (
	"sync"

	"github.com/snapcore/snapd/asserts"
	"github.com/snapcore/snapd/client"
)

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

//return current model assertion
func (a *ClientAdapter) Remodel(b []byte) (string, error) {
	return a.snapdClient.Remodel(b)
}

//return current model assertion
func (a *ClientAdapter) CurrentModelAssertion() (*asserts.Model, error) {
	return a.snapdClient.CurrentModelAssertion()
}

// Snap returns the most recently published revision of the snap with the
// provided name.
func (a *ClientAdapter) Snap(name string) (*client.Snap, *client.ResultInfo, error) {
	return a.snapdClient.Snap(name)
}

//
func (a *ClientAdapter) Start(names []string, opts client.StartOptions) (changeID string, err error) {
	return a.snapdClient.Start(names, opts)
}

// List returns the list of all snaps installed on the system
// with names in the given list; if the list is empty, all snaps.
func (a *ClientAdapter) List(names []string, opts *client.ListOptions) ([]*client.Snap, error) {
	return a.snapdClient.List(names, opts)
}

func (a *ClientAdapter) Find(opts *client.FindOptions) ([]*client.Snap, *client.ResultInfo, error) {
	return a.snapdClient.Find(opts)
}

func (a *ClientAdapter) CreateUser(opts *client.CreateUserOptions) (*client.CreateUserResult, error) {
	return a.snapdClient.CreateUser(opts)
}

func (a *ClientAdapter) RemoveUser(opts *client.RemoveUserOptions) ([]*client.User, error) {
	return a.snapdClient.RemoveUser(opts)
}
