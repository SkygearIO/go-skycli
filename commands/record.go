package commands

import (
	"fmt"
	"os"

	odcontainer "github.com/oursky/ourd-cli/container"
	odrecord "github.com/oursky/ourd-cli/record"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var handleAsset bool
var assetBaseDirectory string
var promptComplexValue bool
var prettyPrint bool
var recordOutputPath string
var createWhenEdit bool
var recordUsePrivateDatabase bool

func usingDatabaseID(c *odcontainer.Container) string {
	if recordUsePrivateDatabase {
		return c.PrivateDatabaseID()
	} else {
		return c.PublicDatabaseID()
	}
}

func newDatabase() *odcontainer.Database {
	c := newContainer()
	return &odcontainer.Database{
		Container:  c,
		DatabaseID: usingDatabaseID(c),
	}
}

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Modify records in database",
	Long:  "record is for modifying records in the database, providing Create, Read, Update and Delete functionality.",
}

var recordImportCmd = &cobra.Command{
	Use:   "import [<path> ...]",
	Short: "Import records to database",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

var recordExportCmd = &cobra.Command{
	Use:   "export <record_id> [<record_id> ...]",
	Short: "Export records from database",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Usage()
			os.Exit(1)
		}
		fmt.Println("not implemented")
	},
}

var recordDeleteCmd = &cobra.Command{
	Use:   "delete <record_id> [<record_id> ...]",
	Short: "Delete Records from database",
	Long:  "Each specified record is deleted from the database.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Usage()
			os.Exit(1)
		}
		fmt.Println("not implemented")
	},
}

var recordSetCmd = &cobra.Command{
	Use:   "set <record_id> <key=value> [<key=value> ...]",
	Short: "Set attributes on a record",
	Run: func(cmd *cobra.Command, args []string) {
		checkMinArgCount(cmd, args, 2)

		modifyRecord, err := odrecord.MakeEmptyRecord(args[0])
		if err != nil {
			fatal(err)
		}

		for _, arg := range args[1:] {
			err := modifyRecord.Assign(arg)
			if err != nil {
				fatal(err)
			}
		}

		db := newDatabase()
		err = db.SaveRecord(modifyRecord)
		if err != nil {
			fatal(err)
		}
	},
}

var recordGetCmd = &cobra.Command{
	Use:   "get <record_id> <key>",
	Short: "Get value of a record attribute",
	Run: func(cmd *cobra.Command, args []string) {
		checkMinArgCount(cmd, args, 2)
		recordID := args[0]
		desiredKey := args[1]
		err := odrecord.CheckRecordID(recordID)
		if err != nil {
			fatal(err)
		}

		db := newDatabase()
		record, err := db.FetchRecord(recordID)
		if err != nil {
			fatal(err)
		}

		desiredValue, err := record.Get(desiredKey)
		if err != nil {
			fatal(err)
		}

		printValue(desiredValue)
	},
}

var recordEditCmd = &cobra.Command{
	Use:   "edit (<record_type|<record_id>)",
	Short: "Edit a record",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			os.Exit(1)
		}
		fmt.Println("not implemented")
	},
}

var recordQueryCmd = &cobra.Command{
	Use:   "query <record_type>",
	Short: "Query records from database",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			os.Exit(1)
		}
		fmt.Println("not implemented")
	},
}

func init() {
	recordCmd.PersistentFlags().BoolVarP(&recordUsePrivateDatabase, "private", "p", false, "Database. Default is Public.")
	viper.BindPFlag("use_private_database", recordCmd.PersistentFlags().Lookup("private"))

	recordImportCmd.Flags().BoolVarP(&handleAsset, "asset", "a", true, "upload assets")
	recordImportCmd.Flags().StringVarP(&assetBaseDirectory, "basedir", "d", "", "base path for locating files to be uploaded")
	recordImportCmd.Flags().BoolVar(&promptComplexValue, "prompt-complex", true, "prompt when complex value is used")

	recordExportCmd.Flags().BoolVarP(&handleAsset, "asset", "a", true, "download assets")
	recordExportCmd.Flags().StringVarP(&assetBaseDirectory, "basedir", "d", "", "base path for locating files to be downloaded")
	recordExportCmd.Flags().BoolVarP(&prettyPrint, "pretty-print", "p", false, "print output in a pretty format")
	recordExportCmd.Flags().StringVarP(&recordOutputPath, "output", "o", "", "Path to save the output to. If not specified, output is printed to stdout with newline delimiter.")
	recordGetCmd.Flags().StringVarP(&recordOutputPath, "output", "o", "", "path to save the output to. If not specified, output is printed to stdout.")
	recordGetCmd.Flags().BoolVarP(&handleAsset, "asset", "a", false, "If value to the key is an asset, download the asset and output the content of the asset.")

	recordEditCmd.Flags().BoolVarP(&createWhenEdit, "new", "n", false, "do not fetch record from database before editing")

	recordQueryCmd.Flags().BoolVarP(&handleAsset, "asset", "a", true, "download assets")
	recordQueryCmd.Flags().StringVarP(&assetBaseDirectory, "basedir", "d", "", "base path for locating files to be downloaded")
	recordQueryCmd.Flags().BoolVarP(&prettyPrint, "pretty-print", "p", false, "print output in a pretty format")
	recordQueryCmd.Flags().StringVarP(&recordOutputPath, "output", "o", "", "Path to save the output to. If not specified, output is printed to stdout with newline delimiter.")

	recordCmd.AddCommand(recordImportCmd)
	recordCmd.AddCommand(recordExportCmd)
	recordCmd.AddCommand(recordDeleteCmd)
	recordCmd.AddCommand(recordSetCmd)
	recordCmd.AddCommand(recordGetCmd)
	recordCmd.AddCommand(recordEditCmd)
	recordCmd.AddCommand(recordQueryCmd)
}
