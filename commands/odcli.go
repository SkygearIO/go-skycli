package commands

import (
	"github.com/oursky/ourd-cli/container"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var OurdCliCmd = &cobra.Command{
	Use:   "odcli",
	Short: "Command line interface to Ourd",
}

var ourdAPIKey string
var ourdEndpoint string
var ourdAccessToken string

func init() {
	OurdCliCmd.PersistentFlags().StringVar(&ourdAPIKey, "api_key", "", "API Key")
	OurdCliCmd.PersistentFlags().StringVar(&ourdEndpoint, "endpoint", "", "Endpoint address")
	OurdCliCmd.PersistentFlags().StringVar(&ourdAccessToken, "access_token", "", "Access token")

	viper.BindPFlag("access_token", OurdCliCmd.PersistentFlags().Lookup("access_token"))
	viper.BindPFlag("endpoint", OurdCliCmd.PersistentFlags().Lookup("endpoint"))
	viper.BindPFlag("api_key", OurdCliCmd.PersistentFlags().Lookup("api_key"))

}

func Execute() {
	viper.SetEnvPrefix("odcli")
	viper.AutomaticEnv()

	viper.SetDefault("endpoint", "http://localhost:3000")

	AddCommands()
	OurdCliCmd.Execute()
}

func AddCommands() {
	OurdCliCmd.AddCommand(recordCmd)
	OurdCliCmd.AddCommand(schemaCmd)
	OurdCliCmd.AddCommand(generateDocCmd)
}

func newContainer() *container.Container {
	return &container.Container{
		APIKey:      viper.GetString("api_key"),
		Endpoint:    viper.GetString("endpoint"),
		AccessToken: viper.GetString("access_token"),
	}
}
