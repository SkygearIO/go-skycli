package commands

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

//var handleAsset bool
//var assetBaseDirectory string
//var promptComplexValue bool
//var prettyPrint bool
//var recordOutputPath string
//var createWhenEdit bool

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Modify schema in database",
	Long:  "The key-value structure and data type of a record type can be modified using this command.",
}

var schemaAddCmd = &cobra.Command{
	Use:   "add <record_type> <column_name> <column_def>",
	Short: "Add a column to the schema of a record type",
	Run: func(cmd *cobra.Command, args []string) {
		checkMinArgCount(cmd, args, 3)
		checkMaxArgCount(cmd, args, 3)

		db := newDatabase()
		err := db.CreateColumn(args[0], args[1], args[2])
		if err != nil {
			fatal(err)
		}
	},
}

var schemaMoveCmd = &cobra.Command{
	Use:     "move <record_type> <column_name> <new_column_name>",
	Short:   "Give a new name to an existing column",
	Aliases: []string{"mv"},
	Run: func(cmd *cobra.Command, args []string) {
		checkMinArgCount(cmd, args, 3)
		checkMaxArgCount(cmd, args, 3)

		db := newDatabase()
		err := db.RenameColumn(args[0], args[1], args[2])
		if err != nil {
			fatal(err)
		}
	},
}

var schemaRemoveCmd = &cobra.Command{
	Use:     "remove <record_type> <column_name>",
	Short:   "Remove a column from the schema of a record type",
	Aliases: []string{"rm", "delete", "del"},
	Run: func(cmd *cobra.Command, args []string) {
		checkMinArgCount(cmd, args, 2)
		checkMaxArgCount(cmd, args, 2)

		db := newDatabase()
		err := db.DeleteColumn(args[0], args[1])
		if err != nil {
			fatal(err)
		}
	},
}

var schemaFetchCmd = &cobra.Command{
	Use:   "fetch <record_type>",
	Short: "Fetch the information of the current record schema",
	Run: func(cmd *cobra.Command, args []string) {
		checkMinArgCount(cmd, args, 0)
		checkMaxArgCount(cmd, args, 0)

		db := newDatabase()
		result, err := db.FetchSchema()
		if err != nil {
			fatal(err)
		}
		printSchemaResult(result)
	},
}

func printSchemaResult(result map[string]interface{}) {
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fatal(err)
	}
	os.Stdout.Write(b)
}

func init() {
	schemaCmd.AddCommand(schemaAddCmd)
	schemaCmd.AddCommand(schemaMoveCmd)
	schemaCmd.AddCommand(schemaRemoveCmd)
	schemaCmd.AddCommand(schemaFetchCmd)
}
