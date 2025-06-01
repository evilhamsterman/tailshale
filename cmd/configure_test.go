package cmd

import (
	"bytes"
	"testing"
	"text/template"

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

func TestWriteIncludeFile(t *testing.T) {
	// Create a temporary filesystem and file path
	appFS := afero.NewMemMapFs()
	path := "/tmp/tailshale_include"

	// Write the include file with a sample command
	tailshaleCommand := "tailshale"
	if err := WriteIncludeFile(appFS, path, tailshaleCommand); err != nil {
		t.Fatalf("Failed to write include file: %v", err)
	}

	// Check if the file was created and contains the expected content
	content, err := afero.ReadFile(appFS, path)
	if err != nil {
		t.Fatalf("Failed to read include file: %v", err)
	}

	var expectedContent bytes.Buffer
	tpl := template.New("includeFileTemplate")
	tpl, err = tpl.Parse(inFile)
	if err != nil {
		t.Fatalf("Error parsing template: %v", err)
	}
	data := struct {
		TailshaleCommand string
	}{
		TailshaleCommand: tailshaleCommand,
	}
	if err := tpl.Execute(&expectedContent, data); err != nil {
		t.Fatalf("Error executing template: %v", err)
	}
	if string(content) != expectedContent.String() {
		t.Errorf("Expected include file content to be %q, but got %q", inFile, string(content))
	}
}
