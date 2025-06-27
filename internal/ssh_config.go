package internal

import (
	"bufio"
	"fmt"
	"slices"
	"strings"

	"github.com/lithammer/dedent"
	"github.com/spf13/afero"
)

const (
	CfgStart                         = "###### Start Tailshale ######"
	CfgEnd                           = "###### End Tailshale ######"
	cfgLocationBeginning cfgLocation = iota
	cfgLocationEnd
	cfgLocationConfig
)

var Cfg = dedent.Dedent(`
# Tailshale SSH configuration
# Do not edit manually. No really

Match exec "%s known-hosts --check %%h"
	KnownHostsCommand %s known-hosts %%h
`)

type cfgLocation int

type SSHConfig struct {
	Beginning string
	Config    string
	End       string
}

func NewSSHConfigFromFile(file afero.File) (*SSHConfig, error) {
	// read the file in lines using a Scanner
	scanner := bufio.NewScanner(file)
	var configLines []string
	var beginningLines []string
	var endLines []string
	l := cfgLocationBeginning
	for scanner.Scan() {
		line := scanner.Text()
		if line == CfgStart {
			l = cfgLocationConfig
			continue
		}
		if line == CfgEnd {
			l = cfgLocationEnd
			continue
		}
		switch l {
		case cfgLocationBeginning:
			beginningLines = append(beginningLines, line)
		case cfgLocationEnd:
			endLines = append(endLines, line)
		case cfgLocationConfig:
			configLines = append(configLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		// Handle the error, e.g., log it or return it
		return nil, fmt.Errorf("error reading SSH config file: %w", err)
	}

	// Create a new SSHConfig instance
	return &SSHConfig{
		Beginning: strings.Join(beginningLines, "\n"),
		End:       strings.Join(endLines, "\n"),
		Config:    strings.Join(configLines, "\n"),
	}, nil
}

// Implement Stringer inteface
func (c SSHConfig) String() string {
	cfgParts := []string{
		c.Beginning,
	}
	if c.Config != "" {
		cfgParts = slices.Concat(cfgParts, []string{
			CfgStart,
			c.Config,
			CfgEnd,
		})
	}
	cfgParts = append(cfgParts, c.End)

	return strings.Join(cfgParts, "\n")
}

var _ fmt.Stringer = SSHConfig{}

// Set the config
func (c *SSHConfig) SetConfig(exePath string) {
	c.Config = fmt.Sprintf(Cfg, exePath, exePath)
}
