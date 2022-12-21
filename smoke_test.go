package tiny

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSmokeTest(t *testing.T) {
	SetupLogger()
	loaded := LoadConfig()

	assert.True(t, loaded, "config should be loaded successfully")
}
