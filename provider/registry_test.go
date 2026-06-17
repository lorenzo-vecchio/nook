package provider

import (
	"context"
	"testing"

	"github.com/lorenzo-vecchio/nook/config"
	"github.com/stretchr/testify/assert"
)

func TestRegistry_Get(t *testing.T) {
	p, ok := Get("vscode")
	assert.True(t, ok)
	assert.NotNil(t, p)

	_, ok = Get("nonexistent")
	assert.False(t, ok)
}

func TestRegistry_List(t *testing.T) {
	names := List()
	assert.Contains(t, names, "vscode")
	assert.Contains(t, names, "chrome")
	assert.Contains(t, names, "dbeaver")
	assert.Contains(t, names, "docker")
	assert.Contains(t, names, "command")
	assert.Len(t, names, 5)
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	provider := &mockProvider{name: "mock"}
	Register(provider)
	defer delete(registry, "mock")

	p, ok := Get("mock")
	assert.True(t, ok)
	assert.Equal(t, provider, p)

	names := List()
	assert.Contains(t, names, "mock")
}

type mockProvider struct {
	name string
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) Detect() (bool, error) {
	return false, nil
}

func (m *mockProvider) Launch(_ context.Context, _ config.Service, _ string, _ map[string]string) error {
	return nil
}
