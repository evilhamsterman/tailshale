//nolint:errcheck
package cmd

import (
	"strings"
	"testing"

	"github.com/evilhamsterman/tailshale/internal"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestAddTailshaeConfig(t *testing.T) {
	sshConfPath := "/tmp/ssh_config"
	cfg := internal.SSHConfig{}
	cfg.SetConfig("tailshale")

	t.Run("CreateConfigFile", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		_ = AddTailshaleConfig(fs, sshConfPath, "tailshale")
		ok, _ := afero.Exists(fs, sshConfPath)
		assert.True(t, ok, "File does not exist")

		ok, _ = afero.FileContainsBytes(fs, sshConfPath, []byte(cfg.Config))
		assert.True(t, ok, "File does not contain a valid config")
	})

	t.Run("AddToExistingFile", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		afero.WriteFile(fs, sshConfPath, []byte("existing\ncontent"), 0644)
		AddTailshaleConfig(fs, sshConfPath, "tailshale")
		ok, _ := afero.FileContainsBytes(fs, sshConfPath, []byte(cfg.Config))
		assert.True(t, ok, "File does not contain a valid config")
	})

	t.Run("UpdateOldConfig", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		oldCfg := strings.TrimSpace(`
		###### Start Tailshale ######
		old config
		###### End Tailshale ######
		`)

		afero.WriteFile(fs, sshConfPath, []byte(oldCfg), 0644)

		AddTailshaleConfig(fs, sshConfPath, "tailshale")
		ok, _ := afero.FileContainsBytes(fs, sshConfPath, []byte(cfg.Config))
		assert.True(t, ok, "New config was not added to the file")

		ok, _ = afero.FileContainsBytes(fs, sshConfPath, []byte(oldCfg))
		assert.False(t, ok, "Old config should not exist in the file")
	})
}

func TestCleanSSHConfig(t *testing.T) {
	sshConfPath := "/tmp/ssh_config"
	fs := afero.NewMemMapFs()
	cfg := internal.SSHConfig{
		Beginning: "Hostname example.com\n  User user\n",
	}
	cfg.SetConfig("tailshale")

	t.Run("Remove Config", func(t *testing.T) {
		afero.WriteFile(fs, sshConfPath, []byte(cfg.String()), 0644)

		err := CleanSSHConfig(fs, sshConfPath)
		assert.NoError(t, err, "Error cleaning SSH config")

		ok, _ := afero.Exists(fs, sshConfPath)
		assert.True(t, ok, "SSH config file should still exist")

		content, _ := afero.ReadFile(fs, sshConfPath)
		assert.NotContains(t, string(content), cfg.Config, "Config should be removed from the file")
		assert.Contains(t, string(content), cfg.Beginning, "Beginning part should remain in the file")
	})

	t.Run("File Does Not Exist", func(t *testing.T) {
		err := CleanSSHConfig(fs, "/non/existent/path")
		assert.NoError(t, err, "Error should be nil when cleaning non-existent file")
	})

}
