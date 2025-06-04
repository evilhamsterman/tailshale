package cmd

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestAddIncludeLine(t *testing.T) {
	// Create a temporary filesystem and file path
	appFS := afero.NewMemMapFs()

	line := templateTailshaleLine()

	// Check if the include line is added correctly to the SSH config file
	_ = AddIncludeLine(appFS, "/tmp/ssh_config")
	assert.FileExists(t, "/tmp/ssh_config", "Expected SSH config file to be created")
	ok, _ := afero.FileContainsBytes(appFS, "/tmp/ssh_config", []byte(line))
	assert.Equal(t, ok, true, "Expected include line to be present in the SSH config file")

	// Check if the file appends the include line correctly to an existing file
	_ = afero.WriteFile(appFS, "/tmp/ssh_config", []byte("Existing content\n"), 0600)
	_ = AddIncludeLine(appFS, "/tmp/ssh_config")
	ok, _ = afero.FileContainsBytes(appFS, "/tmp/ssh_config", []byte(line))
	assert.Equal(t, ok, true, "Expected include line to be present in the SSH config file after appending")
}
