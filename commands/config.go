package commands

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

type config struct {
	AccessToken string `mapstructure:"access_token"`
	APIKey      string `mapstructure:"api_key"`
	Endpoint    string `mapstructure:"endpoint"`
}

var Config config

func loadDefaultConfig() {
	viper.SetDefault("endpoint", "http://localhost:3000")
}

func defaultConfigLocation() string {
	path, err := homedir.Expand("~/.skycli/config.toml")
	if err != nil {
		fatal(err)
	}
	return path
}

func LoadConfigFile() {
	loadDefaultConfig()

	configFile := viper.GetString("config")
	// If config is not set in flag, try the default config location
	if configFile == "" {
		defaultConfigFile := defaultConfigLocation()

		if _, err := os.Stat(defaultConfigFile); err == nil {
			configFile = defaultConfigFile
		}
	}

	if configFile != "" {
		viper.SetConfigFile(configFile)
		err := viper.ReadInConfig()
		if err != nil {
			fatal(fmt.Errorf("Unable to read config file: %s \n", err))
		}
	}

	err := viper.Unmarshal(&Config)
	if err != nil {
		fatal(err)
	}
}
