package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig("logs")
	assert.Error(t, err)
	assert.NotNil(t, config)
}
