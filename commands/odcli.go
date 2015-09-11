package commands

import "github.com/spf13/cobra"

var OurdCliCmd = &cobra.Command{
	Use:   "odcli",
	Short: "Command line interface to Ourd",
}

func init() {

}

func Execute() {
	AddCommands()
	OurdCliCmd.Execute()
}

func AddCommands() {
	OurdCliCmd.AddCommand(recordCmd)
	OurdCliCmd.AddCommand(schemaCmd)
	OurdCliCmd.AddCommand(generateDocCmd)
}
