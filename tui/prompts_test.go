package tui

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockPrompter struct {
	selectFn       func(label string, options []string, defaultOption string) (string, error)
	multiSelectFn  func(label string, options []string, defaults []string) ([]string, error)
	inputFn        func(label string, defaultVal string) (string, error)
	confirmFn      func(label string, defaultVal bool) (bool, error)
}

func (m *mockPrompter) Select(label string, options []string, defaultOption string) (string, error) {
	return m.selectFn(label, options, defaultOption)
}

func (m *mockPrompter) MultiSelect(label string, options []string, defaults []string) ([]string, error) {
	return m.multiSelectFn(label, options, defaults)
}

func (m *mockPrompter) Input(label string, defaultVal string) (string, error) {
	return m.inputFn(label, defaultVal)
}

func (m *mockPrompter) Confirm(label string, defaultVal bool) (bool, error) {
	return m.confirmFn(label, defaultVal)
}

func TestMockSelect(t *testing.T) {
	m := &mockPrompter{
		selectFn: func(_ string, _ []string, _ string) (string, error) {
			return "vscode", nil
		},
	}
	got, err := m.Select("Pick a provider", []string{"vscode", "dbeaver"}, "vscode")
	assert.NoError(t, err)
	assert.Equal(t, "vscode", got)
}

func TestMockMultiSelect(t *testing.T) {
	m := &mockPrompter{
		multiSelectFn: func(_ string, _ []string, _ []string) ([]string, error) {
			return []string{"dev", "staging"}, nil
		},
	}
	got, err := m.MultiSelect("Pick environments", []string{"dev", "staging", "prod"}, []string{"dev"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"dev", "staging"}, got)
}

func TestMockInput(t *testing.T) {
	m := &mockPrompter{
		inputFn: func(_ string, _ string) (string, error) {
			return "my-value", nil
		},
	}
	got, err := m.Input("Enter name", "default")
	assert.NoError(t, err)
	assert.Equal(t, "my-value", got)
}

func TestMockConfirm(t *testing.T) {
	m := &mockPrompter{
		confirmFn: func(_ string, _ bool) (bool, error) {
			return true, nil
		},
	}
	got, err := m.Confirm("Are you sure?", false)
	assert.NoError(t, err)
	assert.True(t, got)
}

func TestMockSelectError(t *testing.T) {
	m := &mockPrompter{
		selectFn: func(_ string, _ []string, _ string) (string, error) {
			return "", errors.New("interrupted")
		},
	}
	_, err := m.Select("Pick", []string{"a", "b"}, "a")
	assert.Error(t, err)
}

func TestSurveyPrompterCompiles(t *testing.T) {
	p := NewPrompter()
	assert.NotNil(t, p)
}
