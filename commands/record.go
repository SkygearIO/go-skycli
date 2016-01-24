package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	skycontainer "github.com/oursky/skycli/container"
	skyrecord "github.com/oursky/skycli/record"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/twinj/uuid"
)

var skipAsset bool
var assetBaseDirectory string
var promptComplexValue bool
var prettyPrint bool
var recordOutputPath string
var createWhenEdit bool
var recordUsePrivateDatabase bool

func usingDatabaseID(c *skycontainer.Container) string {
	if recordUsePrivateDatabase {
		return c.PrivateDatabaseID()
	}
	return c.PublicDatabaseID()
}

func newDatabase() *skycontainer.Database {
	c := newContainer()
	return &skycontainer.Database{
		Container:  c,
		DatabaseID: usingDatabaseID(c),
	}
}

func formatRecordError(err skycontainer.SkygearError) error {
	var fmtError error
	if err.ID != "" {
		fmtError = fmt.Errorf("Record %s: %s", err.ID, err.Message)
	} else {
		fmtError = errors.New(err.Message)
	}
	return fmtError
}

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Modify records in database",
	Long:  "record is for modifying records in the database, providing Create, Read, Update and Delete functionality.",
}

// getRecordList return a generator of all records in the given input stream
func getRecordList(r io.Reader) <-chan *skyrecord.Record {
	c := make(chan *skyrecord.Record)

	go func() {
		defer close(c)

		dec := json.NewDecoder(r)
		for {
			var data map[string]interface{}
			if err := dec.Decode(&data); err == io.EOF {
				break
			} else if err != nil {
				warn(err)
				break
			}

			record, err := skyrecord.MakeRecord(data)
			if err != nil {
				warn(err)
				continue
			}

			c <- record
		}
	}()

	return c
}

func getFileMode(path string) (os.FileMode, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return 0, err
	}

	return info.Mode(), nil
}

// getImportPathList return a generator of all json file in the given path
func getImportPathList(rootPath string) <-chan string {
	c := make(chan string)

	go func() {
		defer close(c)

		filemode, err := getFileMode(rootPath)
		if err != nil {
			warn(err)
			return
		}

		if filemode.IsDir() {
			// Directory
			filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
				matched, err := filepath.Match("*.json", info.Name())
				if err != nil {
					warn(err)
					return nil
				}
				if matched {
					c <- path
				}
				return nil
			})
		} else if filemode.IsRegular() {
			// Single File
			c <- rootPath
		}
	}()

	return c
}

var (
	uploadAssetRegexp   = regexp.MustCompile("^@file:")
	downloadAssetRegexp = regexp.MustCompile("^@asset:")
)

// upload or skip those assets in a record
func uploadAssets(db skycontainer.SkyDB, record *skyrecord.Record, recordDir string) error {
	for idx, val := range record.Data {
		valStr, ok := val.(string)
		if !ok {
			continue
		}

		if uploadAssetRegexp.MatchString(valStr) {
			if skipAsset {
				delete(record.Data, idx)
			} else {
				path := uploadAssetRegexp.ReplaceAllString(valStr, "")
				if !filepath.IsAbs(path) {
					if assetBaseDirectory != "" {
						path = assetBaseDirectory + "/" + path
					} else if recordDir != "" {
						path = recordDir + "/" + path
					}
				}
				assetID, err := db.SaveAsset(path)
				if err != nil {
					return err
				}
				record.Data[idx] = "@asset:" + assetID
			}
		}
	}
	return nil
}

// download or skip those assets in a record
func downloadAssets(db skycontainer.SkyDB, record *skyrecord.Record) error {
	for idx, val := range record.Data {
		valStr, ok := val.(string)
		if !ok {
			continue
		}

		if downloadAssetRegexp.MatchString(valStr) {
			assetID := downloadAssetRegexp.ReplaceAllString(valStr, "")
			assetData, err := db.FetchAsset(assetID)
			if err != nil {
				return err
			}

			var assetPath string
			if assetBaseDirectory == "" {
				assetPath = assetID
			} else {
				err := os.MkdirAll(assetBaseDirectory, 0755)
				if err != nil {
					return err
				}
				assetPath = assetBaseDirectory + "/" + assetID
			}

			err = ioutil.WriteFile(assetPath, assetData, 0644)
			if err != nil {
				fatal(err)
			}

			record.Data[idx] = fmt.Sprintf("@file:%s", assetPath)
		}
	}
	return nil
}

// Show prompt about converting complex value
func complexValueConfirmation(target string) (bool, error) {
	if !promptComplexValue {
		return true, nil
	}

	var response string
	fmt.Printf("Found complex value %s. Convert? (y or n) ", target)
	_, err := fmt.Scanln(&response)
	if err != nil {
		return false, err
	}

	if len(response) == 0 {
		return false, err
	}

	if response[0] == 'y' || response[0] == 'Y' {
		return true, nil
	} else if response[0] == 'n' || response[0] == 'N' {
		return false, nil
	} else {
		fmt.Println("Unexpected response")
		return complexValueConfirmation(target)
	}
}

// Convert those fields with complex value to the corresponding structure
func convertComplexValue(record *skyrecord.Record) error {
	for idx, val := range record.Data {
		valStr, ok := val.(string)
		if !ok {
			continue
		}

		for _, complexType := range ComplexTypeList {
			if complexType.Validate(valStr) {
				convert, err := complexValueConfirmation(valStr)
				if err != nil {
					return err
				}
				if !convert {
					continue
				}

				result, err := complexType.Convert(valStr)
				if err != nil {
					return err
				}
				record.Data[idx] = result
			}
		}
	}
	return nil
}

func importRecord(db skycontainer.SkyDB, record *skyrecord.Record, recordDir string) error {
	err := uploadAssets(db, record, recordDir)
	if err != nil {
		return err
	}

	err = convertComplexValue(record)
	if err != nil {
		return err
	}

	err = db.SaveRecord(record)
	if err != nil {
		return err
	}

	return nil
}

var recordImportCmd = &cobra.Command{
	Use:   "import [<path> ...]",
	Short: "Import records to database",
	Run: func(cmd *cobra.Command, args []string) {
		db := newDatabase()

		// Stdin
		if len(args) == 0 {
			for r := range getRecordList(os.Stdin) {
				err := importRecord(db, r, "")
				if err != nil {
					warn(err)
					continue
				}
			}
		} else {
			for _, path := range args {
				for filename := range getImportPathList(path) {
					f, err := os.Open(filename)
					if err != nil {
						warn(err)
						continue
					}
					recordPath := filepath.Dir(filename)

					for r := range getRecordList(f) {
						err := importRecord(db, r, recordPath)
						if err != nil {
							warn(err)
							continue
						}
					}
				}
			}
		}

		fmt.Println("Import DONE")
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

		for _, arg := range args {
			if err := skyrecord.CheckRecordID(arg); err != nil {
				fatal(err)
			}
		}

		c := newContainer()

		request := skycontainer.GenericRequest{}
		request.Payload = map[string]interface{}{
			"database_id": usingDatabaseID(c),
			"ids":         args,
		}

		response, err := c.MakeRequest("record:delete", &request)
		if err != nil {
			fatal(err)
		}

		if response.IsError() {
			requestError := response.Error()
			fatal(errors.New(requestError.Message))
		}

		resultArray, ok := response.Payload["result"].([]interface{})
		if !ok {
			fatal(fmt.Errorf("Unexpected server data."))
		}

		for i := range resultArray {
			resultData, ok := resultArray[i].(map[string]interface{})
			if !ok {
				warn(fmt.Errorf("Encountered unexpected server data."))
			}

			if skycontainer.IsError(resultData) {
				serverError := skycontainer.MakeError(resultData)
				warn(formatRecordError(serverError))
			}
		}
	},
}

var recordSetCmd = &cobra.Command{
	Use:   "set <record_id> <key=value> [<key=value> ...]",
	Short: "Set attributes on a record",
	Run: func(cmd *cobra.Command, args []string) {
		checkMinArgCount(cmd, args, 2)

		modifyRecord, err := skyrecord.MakeEmptyRecord(args[0])
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
		err := skyrecord.CheckRecordID(recordID)
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

func modifyWithEditor(record *skyrecord.Record) error {
	recordBytes, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}

	f, err := ioutil.TempFile("/tmp", "skycli")
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(recordBytes)
	if err != nil {
		return err
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	editorCmd := exec.Command(editor, f.Name())
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr
	err = editorCmd.Run()
	if err != nil {
		return err
	}

	f.Seek(0, 0)

	jsonBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonBytes, record)
	if err != nil {
		return err
	}
	return nil
}

var recordEditCmd = &cobra.Command{
	Use:   "edit (<record_type|<record_id>)",
	Short: "Edit a record",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			os.Exit(1)
		}

		recordID := args[0]
		if strings.Contains(recordID, "/") {
			err := skyrecord.CheckRecordID(recordID)
			if err != nil {
				fatal(err)
			}
		} else {
			recordID = args[0] + "/" + uuid.NewV4().String()
			createWhenEdit = true
		}

		var record *skyrecord.Record
		var err error
		db := newDatabase()
		if createWhenEdit {
			record, _ = skyrecord.MakeEmptyRecord(recordID)
		} else {
			record, err = db.FetchRecord(recordID)
			if err != nil {
				fatal(err)
			}
		}

		err = modifyWithEditor(record)
		if err != nil {
			fatal(err)
		}

		err = db.SaveRecord(record)
		if err != nil {
			fatal(err)
		}

	},
}
var recordQueryCmd = &cobra.Command{
	Use:   "query <record_type>",
	Short: "Query records from database",
	Run: func(cmd *cobra.Command, args []string) {
		checkMinArgCount(cmd, args, 1)

		recordType := args[0]
		if strings.Contains(recordType, "/") {
			fatal(fmt.Errorf("Record type cannot contain '/'."))
		}

		c := newContainer()
		db := newDatabase()

		request := skycontainer.GenericRequest{}
		request.Payload = map[string]interface{}{
			"database_id": usingDatabaseID(c),
			"record_type": recordType,
		}

		response, err := c.MakeRequest("record:query", &request)
		if err != nil {
			fatal(err)
		}

		if response.IsError() {
			requestError := response.Error()
			fatal(errors.New(requestError.Message))
		}

		resultArray, ok := response.Payload["result"].([]interface{})
		if !ok {
			fatal(fmt.Errorf("Unexpected server data."))
		}

		for i := range resultArray {
			resultData, ok := resultArray[i].(map[string]interface{})
			if !ok {
				warn(fmt.Errorf("Encountered unexpected server data."))
			}

			if skycontainer.IsError(resultData) {
				serverError := skycontainer.MakeError(resultData)
				warn(formatRecordError(serverError))
				continue
			}

			record, err := skyrecord.MakeRecord(resultData)
			if err != nil {
				warn(err)
				continue
			}

			if !skipAsset {
				err = downloadAssets(db, record)
				if err != nil {
					warn(err)
					continue
				}
			}

			var resultJSON []byte
			if prettyPrint {
				resultJSON, err = json.MarshalIndent(resultData, "", "    ")
			} else {
				resultJSON, err = json.Marshal(resultData)
			}
			if err != nil {
				warn(err)
				continue
			}

			if recordOutputPath == "" {
				fmt.Println(string(resultJSON))
			} else {
				err := ioutil.WriteFile(recordOutputPath, resultJSON, 0644)
				if err != nil {
					warn(err)
					continue
				}
			}

		}
	},
}

func init() {
	recordCmd.PersistentFlags().BoolVarP(&recordUsePrivateDatabase, "private", "p", false, "Database. Default is Public.")
	viper.BindPFlag("use_private_database", recordCmd.PersistentFlags().Lookup("private"))

	recordImportCmd.Flags().BoolVar(&skipAsset, "skip-asset", false, "upload assets")
	recordImportCmd.Flags().StringVarP(&assetBaseDirectory, "basedir", "d", "", "base path for locating files to be uploaded")
	recordImportCmd.Flags().BoolVarP(&promptComplexValue, "no-warn-complex", "i", true, "Ignore complex values conversion warnings.")

	recordExportCmd.Flags().BoolVar(&skipAsset, "skip-asset", false, "download assets")
	recordExportCmd.Flags().StringVarP(&assetBaseDirectory, "basedir", "d", "", "base path for locating files to be downloaded")
	recordExportCmd.Flags().BoolVar(&prettyPrint, "pretty-print", false, "print output in a pretty format")
	recordExportCmd.Flags().StringVarP(&recordOutputPath, "output", "o", "", "Path to save the output to. If not specified, output is printed to stdout with newline delimiter.")
	recordGetCmd.Flags().StringVarP(&recordOutputPath, "output", "o", "", "path to save the output to. If not specified, output is printed to stdout.")
	recordGetCmd.Flags().BoolVar(&skipAsset, "skip-asset", false, "If value to the key is an asset, download the asset and output the content of the asset.")

	recordEditCmd.Flags().BoolVarP(&createWhenEdit, "new", "n", false, "do not fetch record from database before editing")

	recordQueryCmd.Flags().BoolVar(&skipAsset, "skip-asset", false, "download assets")
	recordQueryCmd.Flags().StringVarP(&assetBaseDirectory, "basedir", "d", "", "base path for locating files to be downloaded")
	recordQueryCmd.Flags().BoolVar(&prettyPrint, "pretty-print", false, "print output in a pretty format")
	recordQueryCmd.Flags().StringVarP(&recordOutputPath, "output", "o", "", "Path to save the output to. If not specified, output is printed to stdout with newline delimiter.")

	recordCmd.AddCommand(recordImportCmd)
	recordCmd.AddCommand(recordExportCmd)
	recordCmd.AddCommand(recordDeleteCmd)
	recordCmd.AddCommand(recordSetCmd)
	recordCmd.AddCommand(recordGetCmd)
	recordCmd.AddCommand(recordEditCmd)
	recordCmd.AddCommand(recordQueryCmd)
}
