package pkg

import (
	"testing"

	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
)

func TestDefaultTomlContent(t *testing.T) {
	_, err := toml.Marshal(DefaultTomlContent())
	assert.Nil(t, err)
}
