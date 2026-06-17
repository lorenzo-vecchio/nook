package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderStyle(t *testing.T) {
	s := ProviderStyle()
	assert.NotNil(t, s)
}

func TestPathStyle(t *testing.T) {
	s := PathStyle()
	assert.NotNil(t, s)
}

func TestSuccessMark(t *testing.T) {
	m := SuccessMark()
	assert.Contains(t, m, "✔")
}

func TestErrorMark(t *testing.T) {
	m := ErrorMark()
	assert.Contains(t, m, "✖")
}

func TestLoadingMark(t *testing.T) {
	m := LoadingMark()
	assert.Contains(t, m, "⏳")
}

func TestHeaderStyle(t *testing.T) {
	s := HeaderStyle()
	assert.NotNil(t, s)
}

func TestDimStyle(t *testing.T) {
	s := DimStyle()
	assert.NotNil(t, s)
}

func TestBold(t *testing.T) {
	r := Bold("hello")
	assert.NotEmpty(t, r)
}

func TestCyan(t *testing.T) {
	r := Cyan("hello")
	assert.NotEmpty(t, r)
}

func TestGreen(t *testing.T) {
	r := Green("hello")
	assert.NotEmpty(t, r)
}

func TestRed(t *testing.T) {
	r := Red("hello")
	assert.NotEmpty(t, r)
}

func TestYellow(t *testing.T) {
	r := Yellow("hello")
	assert.NotEmpty(t, r)
}

func TestDim(t *testing.T) {
	r := Dim("hello")
	assert.NotEmpty(t, r)
}
