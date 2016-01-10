package container

import (
	"errors"
	"fmt"
	"mime"
	"os"
	"path/filepath"

	odrecord "github.com/oursky/skycli/record"
)

type Database struct {
	Container  *Container
	DatabaseID string
}

func (d *Database) FetchRecord(recordID string) (record *odrecord.Record, err error) {
	request := GenericRequest{}
	request.Payload = map[string]interface{}{
		"database_id": d.DatabaseID,
		"ids":         []string{recordID},
	}

	response, err := d.Container.MakeRequest("record:fetch", &request)
	if err != nil {
		return
	}

	if response.IsError() {
		requestError := response.Error()
		err = errors.New(requestError.Message)
		return
	}

	resultArray, ok := response.Payload["result"].([]interface{})
	if !ok || len(resultArray) < 1 {
		err = fmt.Errorf("Unexpected server data.")
		return
	}

	resultData, ok := resultArray[0].(map[string]interface{})
	if !ok {
		err = fmt.Errorf("Unexpected server data.")
		return
	}

	if IsError(resultData) {
		serverError := MakeError(resultData)
		err = errors.New(serverError.Message)
		return
	}

	record, err = odrecord.MakeRecord(resultData)
	return
}

func (d *Database) SaveRecord(record *odrecord.Record) (err error) {
	request := GenericRequest{}
	request.Payload = map[string]interface{}{
		"database_id": d.DatabaseID,
		"records":     []odrecord.Record{*record},
	}

	response, err := d.Container.MakeRequest("record:save", &request)
	if err != nil {
		return
	}

	if response.IsError() {
		requestError := response.Error()
		err = errors.New(requestError.Message)
		return
	}

	resultArray, ok := response.Payload["result"].([]interface{})
	if !ok || len(resultArray) < 1 {
		err = fmt.Errorf("Unexpected server data.")
		return
	}

	resultData, ok := resultArray[0].(map[string]interface{})
	if !ok {
		err = fmt.Errorf("Unexpected server data.")
		return
	}

	if IsError(resultData) {
		serverError := MakeError(resultData)
		err = errors.New(serverError.Message)
		return
	}
	return
}

func (d *Database) SaveAsset(path string) (assetID string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	//TODO: check whether path is absolute
	info, err := f.Stat()
	if err != nil {
		return
	}
	filename := info.Name()

	//TODO: Use other library to read mime type from content
	filetype := mime.TypeByExtension(filepath.Ext(path))

	response, err := d.Container.MakeAssetRequest("PUT", filename, filetype, f)
	if err != nil {
		return
	}

	if response.IsError() {
		requestError := response.Error()
		err = errors.New(requestError.Message)
		return
	}

	resultData, ok := response.Payload["result"].(map[string]interface{})
	if !ok {
		err = fmt.Errorf("Unexpected server data.")
		return
	}

	assetID, ok = resultData["$name"].(string)
	if !ok {
		err = fmt.Errorf("Unexpected server data.")
		return
	}

	return
}
