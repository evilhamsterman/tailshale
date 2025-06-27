package internal

import (
	"testing"

	"github.com/lithammer/dedent"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSSHConfigFromFile(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected SSHConfig
	}{
		{
			name:  "Empty file",
			input: "",
			expected: SSHConfig{
				Beginning: "",
				End:       "",
				Config:    "",
			},
		},
		{
			name: "File with only Tailshale config",
			input: dedent.Dedent(`
			###### Start Tailshale ######
			config line 1
			config line 2
			###### End Tailshale ######`),
			expected: SSHConfig{
				Beginning: "",
				End:       "",
				Config:    "config line 1\nconfig line 2",
			},
		},
		{
			name: "File with content above the Tailshale config",
			input: dedent.Dedent(`
			Host example.com
			  Hostname example.com
			  User user
			
			###### Start Tailshale ######
			config line 1
			config line 2
			###### End Tailshale ######`),
			expected: SSHConfig{
				Beginning: dedent.Dedent(`
				Host example.com
				  Hostname example.com
				  User user
				  `),
				End:    "",
				Config: "config line 1\nconfig line 2",
			},
		},
		{
			name: "File with content after the Tailshale config",
			input: dedent.Dedent(`
			###### Start Tailshale ######
			config line 1
			config line 2
			###### End Tailshale ######

			Host example.com
			  Hostname example.com
			  User user`),
			expected: SSHConfig{
				Beginning: "",
				End: dedent.Dedent(`
				Host example.com
				  Hostname example.com
				  User user`),
				Config: "config line 1\nconfig line 2",
			},
		},
		{
			name: "File with content before and after the Tailshale config",
			input: dedent.Dedent(`
			Host example.com
			  Hostname example.com
			  User user

			###### Start Tailshale ######
			config line 1
			config line 2
			###### End Tailshale ######

			Host another.com
			  Hostname another.com
			  User anotheruser`),
			expected: SSHConfig{
				Beginning: dedent.Dedent(`
				Host example.com
				  Hostname example.com
				  User user
				  `),
				End: dedent.Dedent(`
				Host another.com
				  Hostname another.com
				  User anotheruser`),
				Config: "config line 1\nconfig line 2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			afero.WriteFile(fs, "/tmp/ssh_config", []byte(tt.input), 0644)

			file, _ := fs.Open("/tmp/ssh_config")

			config, err := NewSSHConfigFromFile(file)
			require.NoError(t, err, "Error creating SSHConfig from file")
			require.NotNil(t, config, "Config should not be nil")
			assert.Equal(t, tt.expected.Beginning, config.Beginning, "Beginning part mismatch")
			assert.Equal(t, tt.expected.End, config.End, "End part mismatch")
			assert.Equal(t, tt.expected.Config, config.Config, "Config part mismatch")
		})
	}
}

func TestSSHConfig_String(t *testing.T) {
	tests := []struct {
		name     string
		config   SSHConfig
		expected string
	}{
		{
			name: "Config with only Tailshale config",
			config: SSHConfig{
				Beginning: "",
				End:       "",
				Config:    "config line 1\nconfig line 2",
			},
			expected: dedent.Dedent(`
			###### Start Tailshale ######
			config line 1
			config line 2
			###### End Tailshale ######
			`),
		},
		{
			name: "Config with content above and below Tailshale config",
			config: SSHConfig{
				Beginning: dedent.Dedent(`
				Host example.com
				  Hostname example.com
				  User user
				`),
				End: dedent.Dedent(`
				Host another.com
				  Hostname another.com
				  User anotheruser
				`),
				Config: "config line 1\nconfig line 2",
			},
			expected: dedent.Dedent(`
			Host example.com
			  Hostname example.com
			  User user

			###### Start Tailshale ######
			config line 1
			config line 2
			###### End Tailshale ######

			Host another.com
			  Hostname another.com
			  User anotheruser
			`),
		},
		{
			name: "No Tailshale config",
			config: SSHConfig{
				Beginning: dedent.Dedent(`
				Host example.com
				  Hostname example.com
				  User user`),
				End: dedent.Dedent(`
				Host another.com
				  Hostname another.com
				  User anotheruser
				`),
				Config: "",
			},
			expected: dedent.Dedent(`
			Host example.com
			  Hostname example.com
			  User user
			
			Host another.com
			  Hostname another.com
			  User anotheruser
			`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.config.String(), "String representation mismatch")
		})
	}
}
