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
var forceConvertComplexValue bool
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

// download those assets in a record
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

			record.Data[idx] = "@file:" + assetPath
		}
	}
	return nil
}

// Show prompt about converting complex value
func complexValueConfirmation(target string) (bool, error) {
	if forceConvertComplexValue {
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

// saveRecord save record at recordDir to db
func saveRecord(db skycontainer.SkyDB, record *skyrecord.Record, recordDir string) error {
	err := record.PreUploadValidate()
	if err != nil {
		return err
	}

	err = uploadAssets(db, record, recordDir)
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

// fetchRecord get the record with recordID from db
func fetchRecord(db skycontainer.SkyDB, recordID string) (*skyrecord.Record, error) {
	err := skyrecord.CheckRecordID(recordID)
	if err != nil {
		return nil, err
	}

	record, err := db.FetchRecord(recordID)
	if err != nil {
		return nil, err
	}

	err = record.PostDownloadHandle()
	if err != nil {
		return nil, err
	}

	if !skipAsset {
		err = downloadAssets(db, record)
		if err != nil {
			return nil, err
		}
	}

	return record, nil
}

// fetchRecord get a record list with recordType from db
func queryRecord(db skycontainer.SkyDB, recordType string) ([]*skyrecord.Record, error) {
	recordList, err := db.QueryRecord(recordType)
	if err != nil {
		return nil, err
	}

	for idx := range recordList {
		err = recordList[idx].PostDownloadHandle()

		if !skipAsset {
			err = downloadAssets(db, recordList[idx])
			if err != nil {
				warn(err)
				continue
			}
		}
	}

	return recordList, nil
}

// printRecordList print the record list to outputFile.
// It would print to stdout if outputFile is not provided.
func printRecordList(recordList []*skyrecord.Record) (err error) {
	var outputFile *os.File
	if recordOutputPath == "" {
		outputFile = os.Stdout
	} else {
		outputFile, err = os.OpenFile(recordOutputPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer outputFile.Close()
	}

	for _, record := range recordList {
		var resultBytes []byte
		if prettyPrint {
			resultBytes, err = record.PrettyPrintBytes()
		} else {
			resultBytes, err = json.Marshal(record)
		}
		if err != nil {
			warn(err)
			continue
		}

		_, err = outputFile.Write(resultBytes)
		if err != nil {
			warn(err)
			continue
		}

		_, err = outputFile.WriteString("\n")
		if err != nil {
			warn(err)
			continue
		}
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
				err := saveRecord(db, r, "")
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
					defer f.Close()

					recordPath := filepath.Dir(filename)

					for r := range getRecordList(f) {
						err := saveRecord(db, r, recordPath)
						if err != nil {
							warn(err)
							continue
						}
					}
				}
			}
		}
	},
}

var recordExportCmd = &cobra.Command{
	Use:   "export <record_id> [<record_id> ...]",
	Short: "Export records from database",
	Run: func(cmd *cobra.Command, args []string) {
		checkMinArgCount(cmd, args, 1)

		db := newDatabase()

		var recordList []*skyrecord.Record
		for _, recordID := range args {
			record, err := fetchRecord(db, recordID)
			if err != nil {
				warn(err)
				continue
			}

			recordList = append(recordList, record)
		}

		printRecordList(recordList)
	},
}

var recordDeleteCmd = &cobra.Command{
	Use:   "delete <record_id> [<record_id> ...]",
	Short: "Delete Records from database",
	Long:  "Each specified record is deleted from the database.",
	Run: func(cmd *cobra.Command, args []string) {
		checkMinArgCount(cmd, args, 1)

		db := newDatabase()

		for _, arg := range args {
			if err := skyrecord.CheckRecordID(arg); err != nil {
				fatal(err)
			}
		}

		err := db.DeleteRecord(args)
		if err != nil {
			fatal(err)
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
		err = saveRecord(db, modifyRecord, "")
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
		checkMaxArgCount(cmd, args, 2)

		db := newDatabase()
		recordID := args[0]
		desiredKey := args[1]

		record, err := fetchRecord(db, recordID)
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

func modifyWithEditor(record *skyrecord.Record) (*skyrecord.Record, error) {
	recordBytes, err := record.PrettyPrintBytes()
	if err != nil {
		return nil, err
	}

	f, err := ioutil.TempFile("/tmp", "skycli")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	_, err = f.Write(recordBytes)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	f.Seek(0, 0)

	jsonBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal(jsonBytes, &data)
	if err != nil {
		return nil, err
	}

	newRecord, err := skyrecord.MakeRecord(data)
	return newRecord, err
}

var recordEditCmd = &cobra.Command{
	Use:   "edit (<record_type|<record_id>)",
	Short: "Edit a record",
	Run: func(cmd *cobra.Command, args []string) {
		checkMinArgCount(cmd, args, 1)
		checkMaxArgCount(cmd, args, 1)

		db := newDatabase()
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
		if createWhenEdit {
			record, _ = skyrecord.MakeEmptyRecord(recordID)
		} else {
			record, err = fetchRecord(db, recordID)
			if err != nil {
				fatal(err)
			}
		}

		record, err = modifyWithEditor(record)
		if err != nil {
			fatal(err)
		}

		err = saveRecord(db, record, "")
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
		checkMaxArgCount(cmd, args, 1)

		db := newDatabase()
		recordType := args[0]
		if strings.Contains(recordType, "/") {
			fatal(fmt.Errorf("Record type cannot contain '/'."))
		}

		recordList, err := queryRecord(db, recordType)
		if err != nil {
			fatal(err)
		}

		err = printRecordList(recordList)
		if err != nil {
			fatal(err)
		}
	},
}

func init() {
	recordCmd.PersistentFlags().BoolVarP(&recordUsePrivateDatabase, "private", "p", false, "Database. Default is Public.")
	viper.BindPFlag("use_private_database", recordCmd.PersistentFlags().Lookup("private"))

	recordImportCmd.Flags().BoolVar(&skipAsset, "skip-asset", false, "Do not upload assets")
	recordImportCmd.Flags().StringVarP(&assetBaseDirectory, "basedir", "d", "", "Base path for locating asset files to be uploaded")
	recordImportCmd.Flags().BoolVarP(&forceConvertComplexValue, "no-warn-complex", "i", false, "Ignore complex values conversion warnings and convert automatically.")

	recordExportCmd.Flags().BoolVar(&skipAsset, "skip-asset", false, "download assets")
	recordExportCmd.Flags().StringVarP(&assetBaseDirectory, "basedir", "d", "", "Base path for asset files to be downloaded")
	recordExportCmd.Flags().BoolVar(&prettyPrint, "pretty-print", false, "Print output in a pretty format")
	recordExportCmd.Flags().StringVarP(&recordOutputPath, "output", "o", "", "Path to save the output to. If not specified, output is printed to stdout with newline delimiter.")

	recordSetCmd.Flags().BoolVar(&skipAsset, "skip-asset", false, "Do not upload assets")
	recordSetCmd.Flags().StringVarP(&assetBaseDirectory, "basedir", "d", "", "Base path for locating files to be uploaded")
	recordSetCmd.Flags().BoolVarP(&forceConvertComplexValue, "no-warn-complex", "i", false, "Ignore complex values conversion warnings and convert automatically.")

	recordGetCmd.Flags().StringVarP(&assetBaseDirectory, "basedir", "d", "", "Base path for asset files to be downloaded.")
	recordGetCmd.Flags().BoolVar(&skipAsset, "skip-asset", false, "Do not download asset.")

	recordEditCmd.Flags().BoolVarP(&createWhenEdit, "new", "n", false, "Do not fetch record from database before editing")

	recordQueryCmd.Flags().BoolVar(&skipAsset, "skip-asset", false, "Do not download assets")
	recordQueryCmd.Flags().StringVarP(&assetBaseDirectory, "basedir", "d", "", "Base path for asset files to be downloaded")
	recordQueryCmd.Flags().BoolVar(&prettyPrint, "pretty-print", false, "Print output in a pretty format")
	recordQueryCmd.Flags().StringVarP(&recordOutputPath, "output", "o", "", "Path to save the output to. If not specified, output is printed to stdout with newline delimiter.")

	recordCmd.AddCommand(recordImportCmd)
	recordCmd.AddCommand(recordExportCmd)
	recordCmd.AddCommand(recordDeleteCmd)
	recordCmd.AddCommand(recordSetCmd)
	recordCmd.AddCommand(recordGetCmd)
	recordCmd.AddCommand(recordEditCmd)
	recordCmd.AddCommand(recordQueryCmd)
}
