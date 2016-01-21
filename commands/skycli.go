package commands

import (
	"github.com/oursky/skycli/container"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var SkygearCliCmd = &cobra.Command{
	Use:   "skycli",
	Short: "Command line interface to Skygear",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		LoadConfigFile()
	},
}

var skygearAPIKey string
var skygearEndpoint string
var skygearAccessToken string

func init() {
	SkygearCliCmd.PersistentFlags().String("config", "", "Config file location. Default is $HOME/.skycli/config.toml")
	SkygearCliCmd.PersistentFlags().StringVar(&skygearAPIKey, "api_key", "", "API Key")
	SkygearCliCmd.PersistentFlags().StringVar(&skygearEndpoint, "endpoint", "", "Endpoint address")
	SkygearCliCmd.PersistentFlags().StringVar(&skygearAccessToken, "access_token", "", "Access token")

	viper.BindPFlag("access_token", SkygearCliCmd.PersistentFlags().Lookup("access_token"))
	viper.BindPFlag("endpoint", SkygearCliCmd.PersistentFlags().Lookup("endpoint"))
	viper.BindPFlag("api_key", SkygearCliCmd.PersistentFlags().Lookup("api_key"))
	viper.BindPFlag("config", SkygearCliCmd.PersistentFlags().Lookup("config"))

}

func Execute() {
	viper.SetEnvPrefix("skycli")
	viper.AutomaticEnv()

	AddCommands()
	SkygearCliCmd.Execute()
}

func AddCommands() {
	SkygearCliCmd.AddCommand(recordCmd)
	SkygearCliCmd.AddCommand(schemaCmd)
	SkygearCliCmd.AddCommand(generateDocCmd)
}

func newContainer() *container.Container {
	return &container.Container{
		APIKey:      Config.APIKey,
		Endpoint:    Config.Endpoint,
		AccessToken: Config.AccessToken,
	}
}
