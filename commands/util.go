package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func checkMinArgCount(cmd *cobra.Command, args []string, min int) {
	if len(args) != min {
		cmd.Usage()
		os.Exit(1)
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	os.Exit(1)
}

func printValue(value interface{}) {
	switch value.(type) {
	case []interface{}:
		data, err := json.Marshal(value)
		if err != nil {
			panic(err)
		}
		fmt.Println("%s\n", data)
	case map[string]interface{}:
		data, err := json.Marshal(value)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", data)
	default:
		fmt.Printf("%v\n", value)
	}
}
