package tiny

import (
	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/hcl"
	"github.com/gookit/config/v2/json"
	"github.com/gookit/config/v2/yamlv3"
	"os"
	"strings"
)

// LoadConfig loads configuration from environment variables and optionally from the specified list of files.
// YAML, JSON and HCL file formats are supported.
func LoadConfig(files ...string) (*config.Config, error) {
	envs := map[string]string{}
	for _, env := range os.Environ() {
		s := strings.SplitN(env, "=", 2)
		envName := s[0]

		envs[envName] = envNameToConfigKey(envName)
	}

	c := config.NewWithOptions("config")
	c.LoadOSEnvs(envs)

	if len(files) > 0 {
		c.AddDriver(yamlv3.Driver)
		c.AddDriver(json.Driver)
		c.AddDriver(hcl.Driver)

		err := c.LoadFiles(files...)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func envNameToConfigKey(envName string) string {
	return strings.ReplaceAll(strings.ToLower(envName), "_", ".")
}
