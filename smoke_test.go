package tiny

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func SmokeTest(t *testing.T) {
	SetupLogger()
	err := LoadConfig()

	assert.Nil(t, err, "config should be loaded successfully")
}
