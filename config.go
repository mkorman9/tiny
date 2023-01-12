package tiny

import (
	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/json"
	"github.com/gookit/config/v2/toml"
	"github.com/gookit/config/v2/yamlv3"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
)

// LoadConfig loads configuration from environment variables and optionally from the specified list of files.
// YAML, JSON and HCL file formats are supported.
// Configuration is stored into global config.Config instance.
func LoadConfig(files ...string) (loaded bool) {
	loaded = true

	if len(files) > 0 {
		config.AddDriver(yamlv3.Driver)
		config.AddDriver(json.Driver)
		config.AddDriver(toml.Driver)

		err := config.LoadFiles(files...)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to load configuration files")
			loaded = false
		}
	}

	envs := map[string]string{}
	for _, env := range os.Environ() {
		s := strings.SplitN(env, "=", 2)
		envName := s[0]

		envs[envName] = envNameToConfigKey(envName)
	}

	config.LoadOSEnvs(envs)
	return
}

func envNameToConfigKey(envName string) string {
	return strings.ReplaceAll(strings.ToLower(envName), "_", ".")
}
