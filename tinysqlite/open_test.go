package tinysqlite

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBasic(t *testing.T) {
	// when
	db, err := Open(":memory:")
	if err != nil {
		assert.Error(t, err, "Open() should not return error")
	}

	tx := db.Raw("SELECT 1")
	if tx.Error != nil {
		assert.Error(t, err, "query should not return error")
	}

	// then
	var result int
	tx.Scan(&result)
	assert.Equal(t, 1, result, "returned result should be correct")
}
