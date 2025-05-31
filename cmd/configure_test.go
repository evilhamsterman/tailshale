package cmd

import (
	"testing"

	"github.com/spf13/afero"
)

func TestAddIncludeLine(t *testing.T) {
	// Create a temporary filesystem and file path
	appFS := afero.NewMemMapFs()

	// Check if the include line is added correctly to the SSH config file
	_ = AddIncludeLine(appFS, "/tmp/ssh_config")
	if ok, _ := afero.FileContainsBytes(appFS, "/tmp/ssh_config", []byte(inFileLine)); !ok {
		t.Errorf("Expected include line %q to be added, but it was not found", inFileLine)
	}

	// Check if the file appends the include line correctly to an existing file
	_ = afero.WriteFile(appFS, "/tmp/ssh_config", []byte("Existing content\n"), 0600)
	_ = AddIncludeLine(appFS, "/tmp/ssh_config")
	if ok, _ := afero.FileContainsBytes(appFS, "/tmp/ssh_config", []byte(inFileLine)); !ok {
		t.Errorf("Expected include line %q to be appended, but it was not found", inFileLine)
	}
}
