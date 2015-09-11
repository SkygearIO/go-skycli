package commands

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var generateDocCmd = &cobra.Command{
	Use:    "gendoc <path>",
	Short:  "Generate documentation",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			os.Exit(1)
		}

		generateDocDir := args[0]
		if !strings.HasSuffix(generateDocDir, "/") {
			generateDocDir += "/"
		}

		cobra.GenMarkdownTree(OurdCliCmd, generateDocDir)
	},
}
