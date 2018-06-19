package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SplitCmdString(t *testing.T) {
	cmd := `put   1024  0   1   5  `
	args := SplitCmdString(cmd)
	assert.Equal(t, 5, len(args))
	assert.Equal(t, "put", args[0])
	assert.Equal(t, "1024", args[1])
	assert.Equal(t, "0", args[2])
	assert.Equal(t, "1", args[3])
	//assert.Equal(t, "5", args[4])
}
