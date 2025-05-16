package simple_test

import (
	"testing"

	"github.com/bramburn/go_ntrip/pkg/simple"
	"github.com/stretchr/testify/assert"
)

// TestAdd tests the Add function
func TestAdd(t *testing.T) {
	result := simple.Add(2, 3)
	assert.Equal(t, 5, result, "2 + 3 should equal 5")
}
