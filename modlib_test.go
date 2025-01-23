// modlib
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

package modlib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadModule(t *testing.T) {

	mod, err := LoadModule("itmod/test/reflection.it")
	assert.NoError(t, err)

	assert.Equal(t, "reflection", mod.Title)
}
