package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CompareVersion(t *testing.T) {
	examples := map[[2]string]bool{
		[2]string{"", ""}:                   false,
		[2]string{"v0.1.1", ""}:             false,
		[2]string{"0.1.1", "0.1.1"}:         false,
		[2]string{"0.1.2", "0.1.1"}:         false,
		[2]string{"v1.0.0", "v2.0.0"}:       false,
		[2]string{"v2.0.0", "v1.0.0"}:       true,
		[2]string{"v1.2.3", "v3.2.1"}:       false,
		[2]string{"v3.2.1", "v1.2.3"}:       true,
		[2]string{"v0.0.7", "v0.0.7"}:       true,
		[2]string{"v0.1.7", "v0.2.7"}:       false,
		[2]string{"v1.3.7", "v1.2.5"}:       true,
		[2]string{"v11.3.7", "v12.2.5"}:     false,
		[2]string{"v11.3.7", "v11.2.5"}:     true,
		[2]string{"v45.21.67", "v23.21.59"}: true,
	}
	for k, v := range examples {
		assert.Equal(t, v, CompareVersion(k[0], k[1]), fmt.Sprintf("v1: %s, v2: %s", k[0], k[1]))
	}
}
